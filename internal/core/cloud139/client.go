package cloud139

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

const (
	BaseURL          = "https://yun.139.com"
	UserNjsURL       = "https://user-njs.yun.139.com"
	ShareKdNjsURL    = "https://share-kd-njs.yun.139.com"
	PersonalKdNjsURL = "https://personal-kd-njs.yun.139.com"
	CatalogV1URL     = BaseURL + "/orchestration/personalCloud/catalog/v1.0"
)

type Cloud139 struct {
	account *db.Account
	client  *http.Client
}

func init() {
	core.RegisterDriver("139", func(account *db.Account) core.CloudDrive {
		return NewCloud139(account)
	})
}

func NewCloud139(account *db.Account) *Cloud139 {
	return &Cloud139{
		account: account,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// ─── 辅助工具 ─────────────────────────────────────────────────────────────────

func md5Hash(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// jsEncodeURIComponent 模拟 JS 的 encodeURIComponent
func jsEncodeURIComponent(str string) string {
	t := url.QueryEscape(str)
	t = strings.ReplaceAll(t, "+", "%20")
	t = strings.ReplaceAll(t, "%21", "!")
	t = strings.ReplaceAll(t, "%27", "'")
	t = strings.ReplaceAll(t, "%28", "(")
	t = strings.ReplaceAll(t, "%29", ")")
	t = strings.ReplaceAll(t, "%2A", "*")
	t = strings.ReplaceAll(t, "%7E", "~")
	return t
}

func (c *Cloud139) getPhone() string {
	re := regexp.MustCompile(`1\d{10}`)
	if re.MatchString(c.account.AccountName) {
		return re.FindString(c.account.AccountName)
	}
	auth := c.account.AuthToken
	if auth != "" {
		token := auth
		if strings.HasPrefix(strings.ToLower(auth), "basic ") {
			token = auth[6:]
		}
		decoded, err := base64.StdEncoding.DecodeString(token)
		if err == nil {
			if match := re.FindString(string(decoded)); match != "" {
				return match
			}
		}
	}
	if c.account.Cookie != "" {
		if match := re.FindString(c.account.Cookie); match != "" {
			return match
		}
	}
	return ""
}

func getNewSignHash(body interface{}, datetime, randomStr string) string {
	s := ""
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		s = jsEncodeURIComponent(string(jsonBytes))
		chars := strings.Split(s, "")
		sort.Strings(chars)
		s = strings.Join(chars, "")
	}
	r := md5Hash(base64.StdEncoding.EncodeToString([]byte(s)))
	c := md5Hash(datetime + ":" + randomStr)
	return strings.ToUpper(md5Hash(r + c))
}

func (c *Cloud139) computeMcloudSign(catalogID string) string {
	now := time.Now().In(time.FixedZone("CST", 8*3600))
	datetime := now.Format("2006-01-02 15:04:05")
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	randomStr := ""
	for i := 0; i < 16; i++ {
		randomStr += string(chars[rand.Intn(len(chars))])
	}

	getDiskBody := map[string]interface{}{
		"catalogID":       catalogID,
		"sortDirection":   1,
		"startNumber":     1,
		"endNumber":       100,
		"filterType":      0,
		"catalogSortType": 0,
		"contentSortType": 0,
		"commonAccountInfo": map[string]interface{}{
			"account":     c.getPhone(),
			"accountType": 1,
		},
	}
	hash := getNewSignHash(getDiskBody, datetime, randomStr)
	return fmt.Sprintf("%s,%s,%s", datetime, randomStr, hash)
}

// ─── HTTP 请求封装 ─────────────────────────────────────────────────────────────

func (c *Cloud139) doRequest(ctx context.Context, method, apiURL string, body interface{}, headers map[string]string) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiURL, bodyReader)
	if err != nil {
		return nil, err
	}

	if c.account.AuthToken != "" {
		auth := c.account.AuthToken
		if !strings.HasPrefix(strings.ToLower(auth), "basic ") {
			auth = "Basic " + auth
		}
		req.Header.Set("Authorization", auth)
	} else if c.account.Cookie != "" {
		req.Header.Set("Cookie", c.account.Cookie)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://yun.139.com/")
	req.Header.Set("Origin", "https://yun.139.com")
	req.Header.Set("x-yun-api-version", "v1")
	req.Header.Set("x-yun-app-channel", "10000034")
	req.Header.Set("x-yun-channel-source", "10000034")
	req.Header.Set("x-yun-client-info", "||9|7.17.2|chrome|143.0.0.0|ff559f01db65afce55f3b4e5d75be4cb||windows 10||zh-CN|||")
	req.Header.Set("x-yun-module-type", "100")
	req.Header.Set("x-yun-svc-type", "1")
	req.Header.Set("mcloud-channel", "1000101")
	req.Header.Set("mcloud-version", "7.17.2")
	req.Header.Set("mcloud-client", "10701")
	req.Header.Set("mcloud-route", "001")

	if strings.Contains(apiURL, "personal-kd-njs") {
		req.Header.Set("INNER-HCY-ROUTER-HTTPS", "1")
		req.Header.Set("x-m4c-caller", "PC")
		req.Header.Set("x-m4c-src", "10002")
		req.Header.Set("x-inner-ntwk", "2")
		req.Header.Set("X-Deviceinfo", "||9|7.17.2|chrome|143.0.0.0|ff559f01db65afce55f3b4e5d75be4cb||windows 10||zh-CN|||")
		req.Header.Set("CMS-DEVICE", "default")
		req.Header.Set("x-huawei-channelSrc", "10000034")
		req.Header.Set("x-SvcType", "1")
	} else if strings.Contains(apiURL, "share-kd-njs") {
		req.Header.Set("caller", "web")
		req.Header.Set("x-m4c-caller", "PC")
		delete(headers, "mcloud-sign")
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 广播响应到仪表盘
	u, _ := url.Parse(apiURL)
	apiPath := apiURL
	if u != nil {
		apiPath = u.Path
	}
	msg := fmt.Sprintf("[139 Debug] 接口 %s 响应: %s", apiPath, string(respBody))
	log.Printf(msg)

	return respBody, nil
}

// ─── CloudDrive 接口实现 ───────────────────────────────────────────────────────

// computeAnySign 根据任意 Body 计算 mcloud-sign
func (c *Cloud139) computeAnySign(body interface{}) string {
	now := time.Now().In(time.FixedZone("CST", 8*3600))
	datetime := now.Format("2006-01-02 15:04:05")
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	randomStr := ""
	for i := 0; i < 16; i++ {
		randomStr += string(chars[rand.Intn(len(chars))])
	}
	hash := getNewSignHash(body, datetime, randomStr)
	return fmt.Sprintf("%s,%s,%s", datetime, randomStr, hash)
}

func (c *Cloud139) GetInfo(ctx context.Context) (*db.Account, error) {
	log.Printf("[139] 正在获取账号信息...")
	headers := map[string]string{
		"caller":         "web",
		"x-m4c-caller":   "PC",
		"mcloud-version": "7.17.2",
		"mcloud-client":  "10701",
	}
	resp, err := c.doRequest(ctx, "POST", UserNjsURL+"/user/getUser", map[string]interface{}{}, headers)
	if err != nil {
		log.Printf("[139] 获取用户信息请求失败: %v", err)
		return nil, err
	}

	var resRaw map[string]interface{}
	if err := json.Unmarshal(resp, &resRaw); err != nil {
		return nil, err
	}

	code := ""
	switch v := resRaw["code"].(type) {
	case string:
		code = v
	case float64:
		code = fmt.Sprintf("%.0f", v)
	}

	if code != "0000" && code != "0" && code != "" {
		if code == "05050009" || code == "1010010003" {
			return nil, fmt.Errorf("登录已失效 (Token Invalid)")
		}
		return nil, fmt.Errorf("API error: %s", code)
	}

	data, _ := resRaw["data"].(map[string]interface{})
	if data == nil {
		data, _ = resRaw["result"].(map[string]interface{})
	}
	if data == nil {
		data = resRaw
	}

	nickname, _ := data["userName"].(string)
	if nickname == "" {
		nickname, _ = data["nickName"].(string)
	}
	if nickname == "" {
		if profile, ok := data["userProfileInfo"].(map[string]interface{}); ok {
			nickname, _ = profile["userName"].(string)
		}
	}

	phone, _ := data["loginName"].(string)
	if phone == "" {
		phone, _ = data["account"].(string)
	}
	if phone == "" {
		phone, _ = data["msisdn"].(string)
	}
	if phone == "" {
		phone, _ = data["phoneNumber"].(string)
	}
	if profile, ok := data["userProfileInfo"].(map[string]interface{}); ok {
		if phone == "" {
			phone, _ = profile["msisdn"].(string)
		}
		if phone == "" {
			phone, _ = profile["loginAccount"].(string)
		}
		if phone == "" {
			phone, _ = profile["account"].(string)
		}
		if phone == "" {
			phone, _ = profile["phoneNumber"].(string)
		}
	}

	userDomainID, _ := data["userDomainId"].(string)

	if nickname == "" {
		nickname = phone
	}
	if nickname == "" {
		nickname = c.account.AccountName
	}
	if nickname == "" {
		nickname = "移动云盘用户"
	}

	c.account.Nickname = nickname
	c.account.Status = 1
	c.account.LastCheck = time.Now()

	rePhone := regexp.MustCompile(`1\d{10}`)
	phoneNum := ""
	if rePhone.MatchString(phone) {
		phoneNum = rePhone.FindString(phone)
	} else {
		phoneNum = c.getPhone()
	}

	if phoneNum != "" {
		c.account.AccountName = phoneNum
	} else if c.account.AccountName == "" {
		c.account.AccountName = nickname
	}

	log.Printf("[139] 账号校验成功: %s (%s)", c.account.Nickname, c.account.AccountName)

	// 1. 尝试从基础信息探测会员 (适配最新等级名称)
	if val, ok := data["userServiceType"].(string); ok && val != "" {
		switch val {
		case "1":
			c.account.VipName = "普通会员"
		case "2":
			c.account.VipName = "白银会员"
		case "3":
			c.account.VipName = "黄金会员"
		case "4":
			c.account.VipName = "钻石会员"
		case "8":
			c.account.VipName = "移动云盘会员"
		default:
			c.account.VipName = "会员类型:" + val
		}
	}

	// 2. 尝试通过权益接口精细化查询 (带签名)
	if phoneNum != "" {
		benefitReq := map[string]interface{}{
			"isNeedBenefit": 1,
			"commonAccountInfo": map[string]interface{}{
				"account":     phoneNum,
				"accountType": 1,
			},
		}

		// 必须计算签名
		benefitSign := c.computeAnySign(benefitReq)
		benefitHeaders := map[string]string{
			"mcloud-sign":    benefitSign,
			"mcloud-channel": "1000101",
		}

		benefitURLs := []string{
			BaseURL + "/orchestration/group-rebuild/member/v1.0/queryUserBenefits",
			BaseURL + "/orchestration/personalCloud/user/v1.0/queryUserBenefits",
		}

		for _, bUrl := range benefitURLs {
			bResp, err := c.doRequest(ctx, "POST", bUrl, benefitReq, benefitHeaders)
			if err != nil {
				continue
			}

			var bRes map[string]interface{}
			if err := json.Unmarshal(bResp, &bRes); err == nil {
				var foundLevel string
				var searchNode func(m map[string]interface{})
				searchNode = func(m map[string]interface{}) {
					if foundLevel != "" {
						return
					}
					for k, v := range m {
						lowerK := strings.ToLower(k)
						if strings.Contains(lowerK, "level") || strings.Contains(lowerK, "vip") || strings.Contains(lowerK, "member") {
							switch val := v.(type) {
							case string:
								if val != "" && val != "0" && val != "普通" && !strings.Contains(val, "http") {
									foundLevel = val
									return
								}
							case float64:
								if val > 0 {
									switch int(val) {
									case 1:
										foundLevel = "普通会员"
									case 2:
										foundLevel = "白银会员"
									case 3:
										foundLevel = "黄金会员"
									case 4:
										foundLevel = "钻石会员"
									default:
										foundLevel = fmt.Sprintf("会员L%d", int(val))
									}
									return
								}
							}
						}
						if subM, ok := v.(map[string]interface{}); ok {
							searchNode(subM)
						} else if subL, ok := v.([]interface{}); ok {
							for _, item := range subL {
								if im, ok := item.(map[string]interface{}); ok {
									searchNode(im)
								}
							}
						}
					}
				}
				searchNode(bRes)

				if foundLevel != "" {
					c.account.VipName = foundLevel
					log.Printf("[139] 成功通过权益接口更新会员等级: %s", c.account.VipName)
					break
				}
			}
		}
	}

	if userDomainID != "" {
		diskReq := map[string]interface{}{"userDomainId": userDomainID}
		diskResp, err := c.doRequest(ctx, "POST", UserNjsURL+"/user/disk/getPersonalDiskInfo", diskReq, headers)
		if err == nil {
			var diskRes struct {
				Data struct {
					DiskSize     string `json:"diskSize"`
					FreeDiskSize string `json:"freeDiskSize"`
				} `json:"data"`
			}
			if json.Unmarshal(diskResp, &diskRes) == nil {
				total, _ := strconv.ParseInt(diskRes.Data.DiskSize, 10, 64)
				free, _ := strconv.ParseInt(diskRes.Data.FreeDiskSize, 10, 64)
				c.account.CapacityTotal = total * 1024 * 1024
				c.account.CapacityUsed = (total - free) * 1024 * 1024
			}
		}
	}

	return c.account, nil
}

func (c *Cloud139) Login(ctx context.Context) error {
	_, err := c.GetInfo(ctx)
	return err
}

func (c *Cloud139) ListFiles(ctx context.Context, parentID string) ([]core.FileInfo, error) {
	if parentID == "" {
		parentID = "/"
	}
	log.Printf("[139] 正在列出目录文件: %s", parentID)
	sign := c.computeMcloudSign(parentID)
	headers := map[string]string{
		"mcloud-sign":            sign,
		"mcloud-version":         "7.17.2",
		"mcloud-channel":         "1000101",
		"mcloud-client":          "10701",
		"INNER-HCY-ROUTER-HTTPS": "1",
	}

	body := map[string]interface{}{
		"pageInfo": map[string]interface{}{
			"pageSize":   100,
			"pageCursor": nil,
		},
		"orderBy":        "updated_at",
		"orderDirection": "DESC",
		"parentFileId":   parentID,
	}

	resp, err := c.doRequest(ctx, "POST", PersonalKdNjsURL+"/hcy/file/list", body, headers)
	if err != nil {
		log.Printf("[139] 列出目录请求失败: %v", err)
		return nil, err
	}

	var res struct {
		Code    string `json:"code"`
		Success bool   `json:"success"`
		Data    struct {
			Items []struct {
				FileID   string `json:"fileId"`
				Name     string `json:"name"`
				Type     string `json:"type"`
				Category string `json:"category"`
				Size     int64  `json:"size"`
				UpdateAt string `json:"updatedAt"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}

	var files []core.FileInfo
	for _, item := range res.Data.Items {
		isFolder := item.Type == "folder" || item.Category == "folder"
		updateTime, _ := time.Parse("2006-01-02 15:04:05", item.UpdateAt)
		files = append(files, core.FileInfo{
			ID:         item.FileID,
			Name:       item.Name,
			Path:       item.FileID,
			IsFolder:   isFolder,
			Size:       item.Size,
			UpdatedAt:  item.UpdateAt,
			UpdateTime: updateTime,
		})
	}
	log.Printf("[139] 目录列出完成: %s, 发现 %d 个项", parentID, len(files))
	return files, nil
}

func (c *Cloud139) CreateFolder(ctx context.Context, parentID, name string) (*core.FileInfo, error) {
	if parentID == "" {
		parentID = "/"
	}
	log.Printf("[139] 正在创建文件夹: Name=%s, ParentID=%s", name, parentID)
	sign := c.computeMcloudSign(parentID)
	headers := map[string]string{
		"mcloud-sign": sign,
	}
	body := map[string]interface{}{
		"parentFileId": parentID,
		"name":         name,
		"type":         "folder",
	}
	resp, err := c.doRequest(ctx, "POST", PersonalKdNjsURL+"/hcy/file/create", body, headers)
	if err != nil {
		log.Printf("[139] 创建文件夹请求失败: %v", err)
		return nil, err
	}
	var res struct {
		Code    string `json:"code"`
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			FileID   string `json:"fileId"`
			ID       string `json:"id"`
			FileName string `json:"fileName"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}
	if !res.Success && res.Code != "0" && res.Code != "0000" && res.Code != "" {
		return nil, fmt.Errorf("139 CreateFolder error [%s]: %s", res.Code, res.Message)
	}

	finalID := res.Data.FileID
	if finalID == "" {
		finalID = res.Data.ID
	}

	log.Printf("[139] 文件夹创建成功: %s", name)
	return &core.FileInfo{
		ID:       finalID,
		Name:     name,
		Path:     finalID,
		IsFolder: true,
	}, nil
}

func (c *Cloud139) ParseShare(ctx context.Context, shareURL, extractCode string) ([]core.FileInfo, error) {
	linkID, passwd, pCaID, err := c.parseShareLink(shareURL)
	if err != nil {
		return nil, err
	}
	if extractCode != "" {
		passwd = extractCode
	}

	info, err := c.getShareInfo(ctx, linkID, passwd, pCaID)
	if err != nil {
		return nil, err
	}

	cst := time.FixedZone("CST", 8*3600)
	var files []core.FileInfo

	// 1. 解析文件夹 (caLst)
	if caLst, ok := info["caLst"].([]interface{}); ok {
		for _, item := range caLst {
			if f, ok := item.(map[string]interface{}); ok {
				// 139 V6 文件夹字段：caName, udTime (20260412155922)
				name, _ := f["caName"].(string)
				udTime, _ := f["udTime"].(string)
				path, _ := f["path"].(string)

				// 时间解析：139 V6 格式通常是 yyyyMMddHHmmss
				var updateTime time.Time
				if len(udTime) == 14 {
					updateTime, _ = time.ParseInLocation("20060102150405", udTime, cst)
				}

				if path == "" {
					caID, _ := f["caID"].(string)
					path = caID
				}

				files = append(files, core.FileInfo{
					ID:         path,
					Name:       name,
					IsFolder:   true,
					UpdatedAt:  updateTime.Format("2006-01-02 15:04:05"),
					UpdateTime: updateTime,
				})
			}
		}
	}

	// 2. 解析文件 (coLst)
	if coLst, ok := info["coLst"].([]interface{}); ok {
		for _, item := range coLst {
			if f, ok := item.(map[string]interface{}); ok {
				// 139 V6 文件字段：coName, udTime, coID
				name, _ := f["coName"].(string)
				udTime, _ := f["udTime"].(string)
				sizeVal, _ := f["size"].(float64)
				path, _ := f["path"].(string)

				var updateTime time.Time
				if len(udTime) == 14 {
					updateTime, _ = time.ParseInLocation("2006-01-02 15:04:05", udTime, cst)
				}

				if path == "" {
					coID, _ := f["coID"].(string)
					path = coID
				}

				files = append(files, core.FileInfo{
					ID:         path,
					Name:       name,
					IsFolder:   false,
					Size:       int64(sizeVal),
					UpdatedAt:  updateTime.Format("2006-01-02 15:04:05"),
					UpdateTime: updateTime,
				})
			}
		}
	}
	return files, nil
}

func (c *Cloud139) SaveFileTo(ctx context.Context, fileID, targetPath string) error {
	return fmt.Errorf("139 driver prefers batch SaveLink operation")
}

func (c *Cloud139) SaveLink(ctx context.Context, shareURL, extractCode, targetPath string, fileIDs []string) error {
	phone := c.getPhone()
	if phone == "" {
		return fmt.Errorf("139 SaveLink error: 无法获取合法的 11 位手机号")
	}

	linkID, passwd, pCaID, err := c.parseShareLink(shareURL)
	if err != nil {
		return err
	}
	if extractCode != "" {
		passwd = extractCode
	}

	info, err := c.getShareInfo(ctx, linkID, passwd, pCaID)
	if err != nil {
		return err
	}

	targetID, err := c.prepareTargetPath(ctx, targetPath)
	if err != nil {
		return err
	}
	if targetID == "/" || targetID == "" {
		targetID = "root"
	}

	idMap := make(map[string]bool)
	for _, id := range fileIDs {
		idMap[id] = true
	}

	coPathLst := []string{}
	if coLst, ok := info["coLst"].([]interface{}); ok {
		for _, item := range coLst {
			if f, ok := item.(map[string]interface{}); ok {
				path, _ := f["path"].(string)
				if path == "" {
					path = fmt.Sprintf("%v/%v", f["parentCatalogID"], f["contentID"])
				}
				if len(fileIDs) == 0 || idMap[path] {
					coPathLst = append(coPathLst, path)
				}
			}
		}
	}

	caPathLst := []string{}
	if caLst, ok := info["caLst"].([]interface{}); ok {
		for _, item := range caLst {
			if f, ok := item.(map[string]interface{}); ok {
				path, _ := f["path"].(string)
				if path == "" {
					caID := f["catalogID"]
					if caID == nil {
						caID = f["caID"]
					}
					path = fmt.Sprintf("%v/%v", f["parentCatalogID"], caID)
				}
				if len(fileIDs) == 0 || idMap[path] {
					caPathLst = append(caPathLst, path)
				}
			}
		}
	}

	if len(coPathLst) == 0 && len(caPathLst) == 0 {
		return nil
	}

	saveBody := map[string]interface{}{
		"createOuterLinkBatchOprTaskReq": map[string]interface{}{
			"msisdn":       phone,
			"ownerAccount": "",
			"taskType":     1,
			"linkID":       linkID,
			"needPassword": passwd != "",
			"taskInfo": map[string]interface{}{
				"linkID":          linkID,
				"needPassword":    passwd != "",
				"contentInfoList": coPathLst,
				"catalogInfoList": caPathLst,
				"newCatalogID":    targetID,
			},
		},
	}

	_, err = c.doRequest(ctx, "POST", ShareKdNjsURL+"/yun-share/richlifeApp/devapp/IBatchOprTask/createOuterLinkBatchOprTask", saveBody, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cloud139) DeleteFile(ctx context.Context, fileID string) error {
	sign := c.computeMcloudSign("/")
	headers := map[string]string{"mcloud-sign": sign, "INNER-HCY-ROUTER-HTTPS": "1"}
	body := map[string]interface{}{"fileIds": []string{fileID}}
	_, err := c.doRequest(ctx, "POST", PersonalKdNjsURL+"/hcy/recyclebin/batchTrash", body, headers)
	return err
}

func (c *Cloud139) parseShareLink(input string) (string, string, string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", "", "", fmt.Errorf("empty share link")
	}
	linkID, passwd, pCaID := "", "", "root"
	urlStr := trimmed
	if !strings.HasPrefix(strings.ToLower(urlStr), "http") {
		urlStr = "https://" + urlStr
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		reBare := regexp.MustCompile(`^[a-zA-Z0-9_-]{4,32}$`)
		if reBare.MatchString(trimmed) {
			return trimmed, "", "root", nil
		}
		return "", "", "", fmt.Errorf("failed to parse url: %v", err)
	}
	q := u.Query()
	linkID = q.Get("linkID")
	if linkID == "" {
		linkID = q.Get("linkId")
	}
	if p := q.Get("pCaID"); p != "" {
		pCaID = p
	}
	if p := q.Get("passwd"); p != "" {
		passwd = p
	} else if p := q.Get("pwd"); p != "" {
		passwd = p
	}
	if linkID == "" && u.Fragment != "" {
		fragment := u.Fragment
		if strings.Contains(fragment, "?") {
			parts := strings.Split(fragment, "?")
			fragment = parts[0]
			fQuery, _ := url.ParseQuery(parts[1])
			linkID = fQuery.Get("linkID")
			if linkID == "" {
				linkID = fQuery.Get("linkId")
			}
			if p := fQuery.Get("pCaID"); p != "" {
				pCaID = p
			}
		}
		if linkID == "" {
			parts := strings.Split(strings.Trim(fragment, "/"), "/")
			if len(parts) > 0 {
				candidate := parts[len(parts)-1]
				reBare := regexp.MustCompile(`^[a-zA-Z0-9_-]{4,32}$`)
				if reBare.MatchString(candidate) {
					linkID = candidate
				}
			}
		}
	}
	if linkID == "" {
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) > 0 {
			candidate := parts[len(parts)-1]
			reBare := regexp.MustCompile(`^[a-zA-Z0-9_-]{4,32}$`)
			if reBare.MatchString(candidate) {
				linkID = candidate
			}
		}
	}
	if linkID == "" {
		return "", "", "", fmt.Errorf("linkID not found in: %s", input)
	}
	return linkID, passwd, pCaID, nil
}

func (c *Cloud139) getShareInfo(ctx context.Context, linkID, passwd, pCaID string) (map[string]interface{}, error) {
	log.Printf("[139] 正在获取分享信息: linkID=%s, pCaID=%s", linkID, pCaID)
	headers := map[string]string{
		"caller": "web", "x-m4c-caller": "PC", "mcloud-client": "10701",
		"mcloud-version": "7.17.2", "mcloud-channel": "1000101",
	}
	body := map[string]interface{}{
		"getOutLinkInfoReq": map[string]interface{}{
			"account": c.getPhone(), "linkID": linkID, "passwd": passwd, "pCaID": pCaID,
			"caSrt": 0, "coSrt": 0, "srtDr": 1, "bNum": 1, "eNum": 200,
		},
	}
	resp, err := c.doRequest(ctx, "POST", ShareKdNjsURL+"/yun-share/richlifeApp/devapp/IOutLink/getOutLinkInfoV6", body, headers)
	if err != nil {
		log.Printf("[139] 请求分享接口失败: %v", err)
		return nil, err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}

	if code, ok := res["code"]; ok {
		codeStr := fmt.Sprintf("%v", code)
		if codeStr != "0" && codeStr != "0000" && codeStr != "" {
			log.Printf("[139] 接口返回错误码: %s, message: %v", codeStr, res["message"])
			return nil, fmt.Errorf("139 getShareInfo error [%s]: %v", codeStr, res["message"])
		}
	}

	// 多节点探测逻辑
	if data, ok := res["data"].(map[string]interface{}); ok {
		return data, nil
	}
	if result, ok := res["result"].(map[string]interface{}); ok {
		return result, nil
	}

	return res, nil
}

func (c *Cloud139) prepareTargetPath(ctx context.Context, path string) (string, error) {
	if path == "" || path == "/" {
		return "root", nil
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	currentID := "root"
	for _, part := range parts {
		files, err := c.ListFiles(ctx, currentID)
		if err != nil {
			return "", err
		}
		found := false
		for _, f := range files {
			if f.IsFolder && f.Name == part {
				currentID = f.ID
				found = true
				break
			}
		}
		if !found {
			newFolder, err := c.CreateFolder(ctx, currentID, part)
			if err != nil {
				return "", err
			}
			currentID = newFolder.ID
		}
	}
	return currentID, nil
}
