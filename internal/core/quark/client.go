package quark

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

const (
	BaseURL    = "https://drive-pc.quark.cn"
	BaseURLApp = "https://drive-m.quark.cn"
	UserAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) quark-cloud-drive/3.14.2 Chrome/112.0.5615.165 Electron/24.1.3.8 Safari/537.36 Channel/pckk_other_ch"
)

type Quark struct {
	account *db.Account
	client  *http.Client
	mparam  map[string]string
}

func init() {
	core.RegisterDriver("quark", func(account *db.Account) core.CloudDrive {
		return NewQuark(account)
	})
}

func NewQuark(account *db.Account) *Quark {
	q := &Quark{
		account: account,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
	q.mparam = q.parseMparam(account.Cookie)
	return q
}

func (q *Quark) parseMparam(cookie string) map[string]string {
	mparam := make(map[string]string)
	reKps := regexp.MustCompile(`(?:^|;| )kps=([a-zA-Z0-9%+/=]+)`)
	reSign := regexp.MustCompile(`(?:^|;| )sign=([a-zA-Z0-9%+/=]+)`)
	reVcode := regexp.MustCompile(`(?:^|;| )vcode=([a-zA-Z0-9%+/=]+)`)

	if match := reKps.FindStringSubmatch(cookie); len(match) > 1 {
		mparam["kps"] = strings.ReplaceAll(match[1], "%25", "%")
	}
	if match := reSign.FindStringSubmatch(cookie); len(match) > 1 {
		mparam["sign"] = strings.ReplaceAll(match[1], "%25", "%")
	}
	if match := reVcode.FindStringSubmatch(cookie); len(match) > 1 {
		mparam["vcode"] = strings.ReplaceAll(match[1], "%25", "%")
	}
	return mparam
}

// ─── HTTP 请求封装 ─────────────────────────────────────────────────────────────

func (q *Quark) doRequest(ctx context.Context, method, apiURL string, query url.Values, body io.Reader, useAppParams bool) ([]byte, error) {
	fullURL := apiURL
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("User-Agent", UserAgent)
	headers.Set("Referer", "https://pan.quark.cn/")

	if useAppParams && q.mparam["kps"] != "" {
		fullURL = strings.Replace(apiURL, BaseURL, BaseURLApp, 1)
		if query == nil {
			query = make(url.Values)
		}
		query.Set("device_model", "M2011K2C")
		query.Set("entry", "default_clouddrive")
		query.Set("_t_group", "0%3A_s_vp%3A1")
		query.Set("dmn", "Mi%2B11")
		query.Set("fr", "android")
		query.Set("pf", "3300")
		query.Set("bi", "35937")
		query.Set("ve", "7.4.5.680")
		query.Set("ss", "411x875")
		query.Set("mi", "M2011K2C")
		query.Set("nt", "5")
		query.Set("nw", "0")
		query.Set("kt", "4")
		query.Set("pr", "ucpro")
		query.Set("sv", "release")
		query.Set("dt", "phone")
		query.Set("data_from", "ucapi")
		query.Set("kps", q.mparam["kps"])
		query.Set("sign", q.mparam["sign"])
		query.Set("vcode", q.mparam["vcode"])
		query.Set("app", "clouddrive")
		query.Set("kkkk", "1")
	} else {
		headers.Set("Cookie", q.account.Cookie)
		if query == nil {
			query = make(url.Values)
		}
		if query.Get("pr") == "" {
			query.Set("pr", "ucpro")
		}
		if query.Get("fr") == "" {
			query.Set("fr", "pc")
		}
	}

	if len(query) > 0 {
		if strings.Contains(fullURL, "?") {
			fullURL += "&" + query.Encode()
		} else {
			fullURL += "?" + query.Encode()
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	req.Header = headers

	resp, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Quark Debug] 请求异常: %s %s, StatusCode=%d, Body=%s", method, fullURL, resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("Quark API HTTP error: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 广播响应到仪表盘
	u, _ := url.Parse(fullURL)
	apiPath := fullURL
	if u != nil {
		apiPath = u.Path
	}
	msg := fmt.Sprintf("[Quark Debug] 接口 %s 响应: %s", apiPath, string(respBody))
	log.Printf(msg)

	return respBody, nil
}

// ─── CloudDrive 接口实现 ───────────────────────────────────────────────────────

func (q *Quark) GetInfo(ctx context.Context) (*db.Account, error) {
	log.Printf("[Quark] 正在获取账号信息...")
	// 预校验 Cookie 格式：PC 网页端接口强制要求包含 __uid
	if !strings.Contains(q.account.Cookie, "__uid=") {
		return nil, fmt.Errorf("夸克网盘 Cookie 格式不正确，缺少核心参数 __uid（请确保获取的是全量网页端 Cookie）")
	}

	apiURL := "https://pan.quark.cn/account/info"
	query := url.Values{}
	query.Set("fr", "pc")
	query.Set("platform", "pc")

	resp, err := q.doRequest(ctx, "GET", apiURL, query, nil, false)
	if err != nil {
		log.Printf("[Quark] 获取账号信息请求失败: %v", err)
		return nil, err
	}

	var resRaw map[string]interface{}
	if err := json.Unmarshal(resp, &resRaw); err != nil {
		return nil, err
	}

	data, ok := resRaw["data"].(map[string]interface{})
	if !ok || data == nil {
		msg, _ := resRaw["message"].(string)
		log.Printf("[Quark] 获取基础信息失败: %v, %s", resRaw["code"], msg)
		return nil, fmt.Errorf("Quark API error: %v, %s", resRaw["code"], msg)
	}

	nickname, _ := data["nickname"].(string)
	if nickname == "" {
		nickname = q.account.AccountName
	}
	if nickname == "" {
		nickname = "Quark User"
	}

	q.account.Nickname = nickname
	q.account.Status = 1
	q.account.LastCheck = time.Now()
	if q.account.AccountName == "" {
		q.account.AccountName = nickname
	}

	log.Printf("[Quark] 基础信息获取成功: %s", nickname)

	// 2. 获取容量和 VIP 信息
	vipFetched := false
	if q.mparam["kps"] != "" {
		log.Printf("[Quark] 发现 kps 参数，尝试调用 App 接口获取容量与会员状态")
		queryGrowth := url.Values{}
		growthResp, err := q.doRequest(ctx, "GET", BaseURLApp+"/1/clouddrive/capacity/growth/info", queryGrowth, nil, true)
		if err == nil && len(growthResp) > 0 {
			var growthRes struct {
				Data struct {
					MemberType    string `json:"member_type"`
					TotalCapacity int64  `json:"total_capacity"`
					UsedCapacity  int64  `json:"used_capacity"`
				} `json:"data"`
			}
			if json.Unmarshal(growthResp, &growthRes) == nil {
				q.account.CapacityTotal = growthRes.Data.TotalCapacity
				q.account.CapacityUsed = growthRes.Data.UsedCapacity
				vipMap := map[string]string{
					"NORMAL":    "普通用户",
					"EXP_SVIP":  "88VIP",
					"SUPER_VIP": "SVIP",
					"Z_VIP":     "SVIP+",
				}
				if name, ok := vipMap[growthRes.Data.MemberType]; ok {
					q.account.VipName = name
				} else if growthRes.Data.MemberType != "" {
					q.account.VipName = growthRes.Data.MemberType
				}
				log.Printf("[Quark] 容量信息 (App): Total=%.2fGB, Used=%.2fGB, MemberType=%s", float64(q.account.CapacityTotal)/1024/1024/1024, float64(q.account.CapacityUsed)/1024/1024/1024, q.account.VipName)
				vipFetched = true
			}
		}
	}

	if !vipFetched {
		log.Printf("[Quark] 尝试调用 Web 探测接口获取容量信息")
		apiURLs := []string{
			"https://pan.quark.cn/1/clouddrive/member?pr=ucpro&fr=pc",
			"https://drive-pc.quark.cn/1/clouddrive/capacity?pr=ucpro&fr=pc",
			"https://pan.quark.cn/1/user/info",
		}

		for _, apiURLWeb := range apiURLs {
			capResp, err := q.doRequest(ctx, "GET", apiURLWeb, nil, nil, false)
			if err != nil || len(capResp) == 0 {
				continue
			}

			var capRaw map[string]interface{}
			if err := json.Unmarshal(capResp, &capRaw); err != nil {
				continue
			}

			dataNode, _ := capRaw["data"].(map[string]interface{})
			metadataNode, _ := capRaw["metadata"].(map[string]interface{})
			resNode := dataNode
			if resNode == nil {
				resNode = metadataNode
			}
			if resNode == nil {
				resNode = capRaw
			}

			if resNode != nil {
				capInfo, _ := resNode["cap_info"].(map[string]interface{})
				if capInfo == nil {
					capInfo = resNode
				}

				total, used := float64(0), float64(0)
				if v, ok := capInfo["total"].(float64); ok {
					total = v
				}
				if v, ok := capInfo["total_capacity"].(float64); ok {
					total = v
				}
				if v, ok := capInfo["cap_total"].(float64); ok {
					total = v
				}
				if v, ok := capInfo["used"].(float64); ok {
					used = v
				}
				if v, ok := capInfo["use_capacity"].(float64); ok {
					used = v
				}
				if v, ok := capInfo["cap_used"].(float64); ok {
					used = v
				}

				if total > 0 {
					q.account.CapacityTotal = int64(total)
					q.account.CapacityUsed = int64(used)
					vipFetched = true
				}

				if mt, ok := resNode["member_type"]; ok {
					vipMap := map[string]string{"NORMAL": "普通用户", "EXP_SVIP": "88VIP", "SUPER_VIP": "SVIP", "Z_VIP": "SVIP+"}
					switch v := mt.(type) {
					case string:
						if name, ok := vipMap[v]; ok {
							q.account.VipName = name
						}
					case float64:
						if v == 0 {
							q.account.VipName = "普通用户"
						} else if v == 1 {
							q.account.VipName = "VIP"
						} else if v == 2 {
							q.account.VipName = "SVIP"
						}
					}
				}

				if vipFetched {
					log.Printf("[Quark] 容量信息 (Web): Total=%.2fGB, Used=%.2fGB", total/1024/1024/1024, used/1024/1024/1024)
					break
				}
			}
		}
	}

	return q.account, nil
}

func (q *Quark) Login(ctx context.Context) error {
	_, err := q.GetInfo(ctx)
	return err
}

func (q *Quark) ListFiles(ctx context.Context, parentID string) ([]core.FileInfo, error) {
	if parentID == "" || parentID == "/" {
		parentID = "0"
	}
	log.Printf("[Quark] 正在列出目录文件: FID=%s", parentID)
	apiURL := BaseURL + "/1/clouddrive/file/sort"
	query := url.Values{}
	query.Set("pdir_fid", parentID)
	query.Set("_page", "1")
	query.Set("_size", "100")
	query.Set("_sort", "file_type:asc,updated_at:desc")
	query.Set("_fetch_total", "1")
	query.Set("fetch_risk_file_name", "1")
	query.Set("pr", "ucpro")
	query.Set("fr", "pc")

	resp, err := q.doRequest(ctx, "GET", apiURL, query, nil, false)
	if err != nil {
		log.Printf("[Quark] 列出目录请求失败: %v", err)
		return nil, err
	}

	var res struct {
		Code interface{} `json:"code"`
		Data struct {
			List []map[string]interface{} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal quark response: %v", err)
	}

	codeStr := fmt.Sprintf("%v", res.Code)
	if codeStr != "0" && codeStr != "0.0" {
		log.Printf("[Quark] 列出目录接口返回错误: %v", res.Code)
		return nil, fmt.Errorf("Quark API error: %v", res.Code)
	}

	var files []core.FileInfo
	for _, item := range res.Data.List {
		fid, _ := item["fid"].(string)
		fileName, _ := item["file_name"].(string)
		dir, _ := item["dir"].(bool)
		sizeVal, _ := item["size"].(float64)
		updateAtVal, _ := item["updated_at"].(float64)

		updateTime := time.Unix(int64(updateAtVal)/1000, 0)
		files = append(files, core.FileInfo{
			ID:         fid,
			Name:       fileName,
			Path:       fid,
			IsFolder:   dir,
			Size:       int64(sizeVal),
			UpdatedAt:  updateTime.Format("2006-01-02 15:04:05"),
			UpdateTime: updateTime,
		})
	}
	log.Printf("[Quark] 目录列出完成: FID=%s, 发现 %d 个项", parentID, len(files))
	return files, nil
}

func (q *Quark) CreateFolder(ctx context.Context, parentID, name string) (*core.FileInfo, error) {
	if parentID == "" || parentID == "/" {
		parentID = "0"
	}
	log.Printf("[Quark] 正在创建文件夹: Name=%s, ParentFID=%s", name, parentID)
	apiURL := BaseURL + "/1/clouddrive/file"
	query := url.Values{}
	query.Set("pr", "ucpro")
	query.Set("fr", "pc")

	body := map[string]interface{}{
		"pdir_fid":  parentID,
		"file_name": name,
		"dir_path":  "",
	}
	jsonBody, _ := json.Marshal(body)
	resp, err := q.doRequest(ctx, "POST", apiURL, query, strings.NewReader(string(jsonBody)), false)
	if err != nil {
		log.Printf("[Quark] 创建文件夹请求失败: %v", err)
		return nil, err
	}

	var res struct {
		Code interface{} `json:"code"`
		Data struct {
			Fid string `json:"fid"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}

	codeStr := fmt.Sprintf("%v", res.Code)
	if codeStr != "0" && codeStr != "0.0" {
		log.Printf("[Quark] 创建文件夹接口返回错误: %v", res.Code)
		return nil, fmt.Errorf("Quark API error: %v", res.Code)
	}
	log.Printf("[Quark] 文件夹创建成功: %s (FID: %s)", name, res.Data.Fid)
	return &core.FileInfo{
		ID:       res.Data.Fid,
		Name:     name,
		Path:     res.Data.Fid,
		IsFolder: true,
	}, nil
}

func (q *Quark) DeleteFile(ctx context.Context, fileID string) error {
	log.Printf("[Quark] 正在删除文件: FID=%s", fileID)
	apiURL := BaseURL + "/1/clouddrive/file/delete"
	body := map[string]interface{}{
		"action_type":  2,
		"filelist":     []string{fileID},
		"exclude_fids": []string{},
	}
	jsonBody, _ := json.Marshal(body)
	_, err := q.doRequest(ctx, "POST", apiURL, nil, strings.NewReader(string(jsonBody)), false)
	return err
}

func (q *Quark) getStoken(ctx context.Context, pwdID, extractCode string) (string, error) {
	tokenURL := BaseURL + "/1/clouddrive/share/sharepage/token"
	tokenBody := map[string]interface{}{
		"pwd_id":   pwdID,
		"passcode": extractCode,
	}
	jsonToken, _ := json.Marshal(tokenBody)
	resp, err := q.doRequest(ctx, "POST", tokenURL, nil, strings.NewReader(string(jsonToken)), true)
	if err != nil {
		return "", err
	}
	var tokenRes struct {
		Code interface{} `json:"code"`
		Data struct {
			Stoken string `json:"stoken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &tokenRes); err != nil {
		return "", err
	}
	codeStr := fmt.Sprintf("%v", tokenRes.Code)
	if codeStr != "0" && codeStr != "0.0" {
		log.Printf("[Quark] 获取 Stoken 失败: %v", tokenRes.Code)
		return "", fmt.Errorf("Quark token error: %v", tokenRes.Code)
	}
	return tokenRes.Data.Stoken, nil
}

func (q *Quark) extractShareParams(shareURL string) (pwdID, pdirFID string) {
	reID := regexp.MustCompile(`/s/(\w+)`)
	if match := reID.FindStringSubmatch(shareURL); len(match) >= 2 {
		pwdID = match[1]
	}
	reFID := regexp.MustCompile(`/share/([a-fA-F0-9]{32})`)
	if match := reFID.FindStringSubmatch(shareURL); len(match) >= 2 {
		pdirFID = match[1]
	}
	if pdirFID == "" {
		pdirFID = "0"
	}
	return
}

func (q *Quark) ParseShare(ctx context.Context, shareURL, extractCode string) ([]core.FileInfo, error) {
	pwdID, pdirFID := q.extractShareParams(shareURL)
	if pwdID == "" {
		return nil, fmt.Errorf("invalid quark share url: %s", shareURL)
	}

	log.Printf("[Quark] 正在解析分享链接: pwdID=%s, pdirFID=%s", pwdID, pdirFID)

	stoken, err := q.getStoken(ctx, pwdID, extractCode)
	if err != nil {
		log.Printf("[Quark] 解析分享链接失败 (Stoken获取失败): %v", err)
		return nil, err
	}

	detailURL := BaseURL + "/1/clouddrive/share/sharepage/detail"
	detailQuery := url.Values{}
	detailQuery.Set("pwd_id", pwdID)
	detailQuery.Set("stoken", stoken)
	detailQuery.Set("pdir_fid", pdirFID)
	detailQuery.Set("pr", "ucpro")
	detailQuery.Set("fr", "pc")
	detailQuery.Set("force", "0")
	detailQuery.Set("_page", "1")
	detailQuery.Set("_size", "100")
	detailQuery.Set("_fetch_total", "1")
	detailQuery.Set("_sort", "file_type:asc,updated_at:desc")
	resp, err := q.doRequest(ctx, "GET", detailURL, detailQuery, nil, true)
	if err != nil {
		log.Printf("[Quark] 解析分享链接失败 (详情请求失败): %v", err)
		return nil, err
	}

	var detailRes struct {
		Data struct {
			List []struct {
				Fid           string `json:"fid"`
				FileName      string `json:"file_name"`
				Dir           bool   `json:"dir"`
				Size          int64  `json:"size"`
				UpdateAt      int64  `json:"updated_at"`
				ShareFidToken string `json:"share_fid_token"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &detailRes); err != nil {
		return nil, err
	}

	var files []core.FileInfo
	for _, item := range detailRes.Data.List {
		updateTime := time.Unix(item.UpdateAt/1000, 0)
		files = append(files, core.FileInfo{
			ID:         item.Fid,
			Name:       item.FileName,
			IsFolder:   item.Dir,
			Size:       item.Size,
			UpdatedAt:  updateTime.Format("2006-01-02 15:04:05"),
			UpdateTime: updateTime,
		})
	}
	log.Printf("[Quark] 解析分享完成: 发现 %d 个项", len(files))
	return files, nil
}

func (q *Quark) SaveFileTo(ctx context.Context, fileID, targetPath string) error {
	return fmt.Errorf("quark driver prefers batch SaveLink operation")
}

func (q *Quark) SaveLink(ctx context.Context, shareURL, extractCode, targetPath string, fileIDs []string) error {
	pwdID, pdirFID := q.extractShareParams(shareURL)
	if pwdID == "" {
		return fmt.Errorf("invalid quark share url: %s", shareURL)
	}

	log.Printf("[Quark] 正在提交转存任务: pwdID=%s, targetPath=%s", pwdID, targetPath)

	stoken, err := q.getStoken(ctx, pwdID, extractCode)
	if err != nil {
		log.Printf("[Quark] 转存失败 (Stoken获取失败): %v", err)
		return err
	}

	detailURL := BaseURL + "/1/clouddrive/share/sharepage/detail"
	detailQuery := url.Values{}
	detailQuery.Set("pwd_id", pwdID)
	detailQuery.Set("stoken", stoken)
	detailQuery.Set("pdir_fid", pdirFID)
	detailQuery.Set("pr", "ucpro")
	detailQuery.Set("fr", "pc")
	detailQuery.Set("_page", "1")
	detailQuery.Set("_size", "100")
	resp, err := q.doRequest(ctx, "GET", detailURL, detailQuery, nil, true)
	if err != nil {
		return err
	}
	var detailRes struct {
		Data struct {
			List []struct {
				Fid           string `json:"fid"`
				ShareFidToken string `json:"share_fid_token"`
			} `json:"list"`
		} `json:"data"`
	}
	json.Unmarshal(resp, &detailRes)

	targetID, err := q.prepareTargetPath(ctx, targetPath)
	if err != nil {
		log.Printf("[Quark] 转存失败 (准备目标路径失败): %v", err)
		return err
	}

	idMap := make(map[string]bool)
	for _, id := range fileIDs {
		parts := strings.Split(id, "|")
		idMap[parts[0]] = true
	}

	var fids []string
	var tokens []string
	for _, item := range detailRes.Data.List {
		if len(fileIDs) == 0 || idMap[item.Fid] {
			fids = append(fids, item.Fid)
			tokens = append(tokens, item.ShareFidToken)
		}
	}

	if len(fids) == 0 {
		log.Printf("[Quark] 转存取消: 未发现匹配的文件 FID")
		return nil
	}

	saveURL := BaseURL + "/1/clouddrive/share/sharepage/save"
	saveBody := map[string]interface{}{
		"fid_list": fids, "fid_token_list": tokens, "to_pdir_fid": targetID,
		"pwd_id": pwdID, "stoken": stoken, "pdir_fid": pdirFID, "scene": "link",
	}
	jsonSave, _ := json.Marshal(saveBody)
	_, err = q.doRequest(ctx, "POST", saveURL, nil, strings.NewReader(string(jsonSave)), true)
	if err != nil {
		log.Printf("[Quark] 转存执行异常: %v", err)
		return err
	}
	log.Printf("[Quark] 转存成功: 已保存 %d 个项至 FID=%s", len(fids), targetID)
	return nil
}

func (q *Quark) prepareTargetPath(ctx context.Context, path string) (string, error) {
	if path == "" || path == "/" {
		return "0", nil
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	currentID := "0"
	for _, part := range parts {
		files, err := q.ListFiles(ctx, currentID)
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
			newFolder, err := q.CreateFolder(ctx, currentID, part)
			if err != nil {
				return "", err
			}
			currentID = newFolder.ID
		}
	}
	return currentID, nil
}
