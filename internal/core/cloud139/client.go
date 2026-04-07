package cloud139

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

const (
	BaseURL           = "https://yun.139.com"
	UserNjsURL        = "https://user-njs.yun.139.com"
	ShareKdNjsURL     = "https://share-kd-njs.yun.139.com"
	PersonalKdNjsURL  = "https://personal-kd-njs.yun.139.com"
	CatalogV1URL      = BaseURL + "/orchestration/personalCloud/catalog/v1.0"
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

func getNewSignHash(body interface{}, datetime, randomStr string) string {
	s := ""
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		s = string(jsonBytes)
		s = url.QueryEscape(s)
		// 139 的 QueryEscape 与 JS 的 encodeURIComponent 有细微差别（空格变+等），
		// 但核心逻辑是字符排序后的 MD5。
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
		"catalogID":     catalogID,
		"sortDirection": 1,
		"startNumber":   1,
		"endNumber":     100,
		"filterType":    0,
		"catalogSortType": 0,
		"contentSortType": 0,
		"commonAccountInfo": map[string]interface{}{
			"account":     c.account.AccountName,
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

	// 默认 Headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://yun.139.com/")
	req.Header.Set("Origin", "https://yun.139.com")

	// 身份认证
	if c.account.AuthToken != "" {
		auth := c.account.AuthToken
		if !strings.HasPrefix(strings.ToLower(auth), "basic ") {
			auth = "Basic " + auth
		}
		req.Header.Set("Authorization", auth)
	} else if c.account.Cookie != "" {
		req.Header.Set("Cookie", c.account.Cookie)
	}

	// 额外 Headers
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

	return respBody, nil
}

// ─── CloudDrive 接口实现 ───────────────────────────────────────────────────────

func (c *Cloud139) GetInfo(ctx context.Context) (*db.Account, error) {
	headers := map[string]string{
		"caller":         "web",
		"x-m4c-caller":   "PC",
		"mcloud-version": "7.17.2",
		"mcloud-client":  "10701",
	}
	resp, err := c.doRequest(ctx, "POST", UserNjsURL+"/user/getUser", map[string]interface{}{}, headers)
	if err != nil {
		return nil, err
	}

	var res struct {
		Code string `json:"code"`
		Data struct {
			UserDomainID string `json:"userDomainId"`
			Nickname     string `json:"userName"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}
	if res.Code != "0000" {
		return nil, fmt.Errorf("API error: %s", res.Code)
	}

	c.account.Nickname = res.Data.Nickname
	c.account.Status = 1
	c.account.LastCheck = time.Now()
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
		"orderBy":         "updated_at",
		"orderDirection":  "DESC",
		"parentFileId":    parentID,
	}

	resp, err := c.doRequest(ctx, "POST", PersonalKdNjsURL+"/hcy/file/list", body, headers)
	if err != nil {
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
		files = append(files, core.FileInfo{
			ID:        item.FileID,
			Name:      item.Name,
			IsFolder:  isFolder,
			Size:      item.Size,
			UpdatedAt: item.UpdateAt,
		})
	}
	return files, nil
}

func (c *Cloud139) CreateFolder(ctx context.Context, name, parentID string) (string, error) {
	if parentID == "" {
		parentID = "/"
	}
	sign := c.computeMcloudSign(parentID)
	headers := map[string]string{
		"mcloud-sign": sign,
		"mcloud-version": "7.17.2",
		"INNER-HCY-ROUTER-HTTPS": "1",
	}
	body := map[string]interface{}{
		"parentFileId": parentID,
		"name":         name,
		"type":         "folder",
	}
	resp, err := c.doRequest(ctx, "POST", PersonalKdNjsURL+"/hcy/file/create", body, headers)
	if err != nil {
		return "", err
	}
	var res struct {
		Code string `json:"code"`
		Data struct {
			FileID string `json:"fileId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return "", err
	}
	return res.Data.FileID, nil
}

func (c *Cloud139) DeleteFile(ctx context.Context, fileID string) error {
	sign := c.computeMcloudSign("/")
	headers := map[string]string{
		"mcloud-sign": sign,
		"INNER-HCY-ROUTER-HTTPS": "1",
	}
	body := map[string]interface{}{
		"fileIds": []string{fileID},
	}
	_, err := c.doRequest(ctx, "POST", PersonalKdNjsURL+"/hcy/recyclebin/batchTrash", body, headers)
	return err
}

// ─── 分享转存相关 ─────────────────────────────────────────────────────────────

func (c *Cloud139) parseShareLink(input string) (string, string, error) {
	u, err := url.Parse(input)
	if err != nil {
		// 尝试直接作为 linkID 处理
		return input, "", nil
	}
	linkID := u.Query().Get("linkID")
	if linkID == "" {
		// 尝试从路径或 Hash 中提取，这里简化处理，实际可能需要更复杂的正则表达式
		parts := strings.Split(u.Path, "/")
		linkID = parts[len(parts)-1]
	}
	passwd := u.Query().Get("passwd")
	if passwd == "" {
		passwd = u.Query().Get("pwd")
	}
	return linkID, passwd, nil
}

func (c *Cloud139) getShareInfo(ctx context.Context, linkID, passwd, pCaID string) (map[string]interface{}, error) {
	headers := map[string]string{
		"caller":         "web",
		"x-m4c-caller":   "PC",
		"mcloud-client":  "10701",
		"mcloud-version": "7.17.2",
	}
	body := map[string]interface{}{
		"getOutLinkInfoReq": map[string]interface{}{
			"account": c.account.AccountName,
			"linkID":  linkID,
			"passwd":  passwd,
			"pCaID":   pCaID,
			"caSrt":   0,
			"coSrt":   0,
			"srtDr":   1,
			"bNum":    1,
			"eNum":    200,
		},
	}
	resp, err := c.doRequest(ctx, "POST", ShareKdNjsURL+"/yun-share/richlifeApp/devapp/IOutLink/getOutLinkInfoV6", body, headers)
	if err != nil {
		return nil, err
	}
	var res struct {
		Code string                 `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Cloud139) SaveLink(ctx context.Context, shareURL, extractCode, targetPath string) error {
	linkID, passwd, err := c.parseShareLink(shareURL)
	if err != nil {
		return err
	}
	if extractCode != "" {
		passwd = extractCode
	}

	// 1. 获取分享内容 (根目录)
	info, err := c.getShareInfo(ctx, linkID, passwd, "root")
	if err != nil {
		return err
	}

	// 2. 获取目标目录 ID (逐级查找或创建)
	targetID, err := c.prepareTargetPath(ctx, targetPath)
	if err != nil {
		return err
	}

	// 3. 构建转存请求
	// 注意：139 转存需要 path 列表，格式为 "parentID/fileID"
	var coPathLst []string
	if coLst, ok := info["coLst"].([]interface{}); ok {
		for _, item := range coLst {
			if f, ok := item.(map[string]interface{}); ok {
				coPathLst = append(coPathLst, fmt.Sprintf("%v/%v", f["parentCatalogID"], f["contentID"]))
			}
		}
	}
	
	var caPathLst []string
	if caLst, ok := info["caLst"].([]interface{}); ok {
		for _, item := range caLst {
			if f, ok := item.(map[string]interface{}); ok {
				caPathLst = append(caPathLst, fmt.Sprintf("%v/%v", f["parentCatalogID"], f["catalogID"]))
			}
		}
	}

	if len(coPathLst) == 0 && len(caPathLst) == 0 {
		return nil // 无内容可存
	}

	saveBody := map[string]interface{}{
		"createOuterLinkBatchOprTaskReq": map[string]interface{}{
			"msisdn":       c.account.AccountName,
			"ownerAccount": "",
			"taskType":     1,
			"linkID":       linkID,
			"needPassword": passwd != "",
			"taskInfo": map[string]interface{}{
				"linkID":            linkID,
				"needPassword":      passwd != "",
				"contentInfoList":   coPathLst,
				"catalogInfoList":   caPathLst,
				"newCatalogID":      targetID,
			},
		},
	}

	headers := map[string]string{
		"caller":         "web",
		"x-m4c-caller":   "PC",
		"mcloud-client":  "10701",
	}
	_, err = c.doRequest(ctx, "POST", ShareKdNjsURL+"/yun-share/richlifeApp/devapp/IBatchOprTask/createOuterLinkBatchOprTask", saveBody, headers)
	return err
}

func (c *Cloud139) prepareTargetPath(ctx context.Context, path string) (string, error) {
	if path == "" || path == "/" {
		return "/", nil
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	currentID := "/"
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
			newID, err := c.CreateFolder(ctx, part, currentID)
			if err != nil {
				return "", err
			}
			currentID = newID
		}
	}
	return currentID, nil
}
