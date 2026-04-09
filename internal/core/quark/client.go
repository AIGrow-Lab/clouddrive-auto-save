package quark

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
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
	if useAppParams && q.mparam["kps"] != "" {
		fullURL = strings.Replace(apiURL, BaseURL, BaseURLApp, 1)
		if query == nil {
			query = make(url.Values)
		}
		query.Set("pr", "ucpro")
		query.Set("fr", "android")
		query.Set("kps", q.mparam["kps"])
		query.Set("sign", q.mparam["sign"])
		query.Set("vcode", q.mparam["vcode"])
	}

	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cookie", q.account.Cookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Referer", "https://pan.quark.cn/")

	resp, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

// ─── CloudDrive 接口实现 ───────────────────────────────────────────────────────

func (q *Quark) GetInfo(ctx context.Context) (*db.Account, error) {
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
		return nil, err
	}

	var resRaw map[string]interface{}
	if err := json.Unmarshal(resp, &resRaw); err != nil {
		return nil, err
	}

	// 只要有 data 节点且不为空，就认为请求成功
	data, ok := resRaw["data"].(map[string]interface{})
	if !ok || data == nil {
		msg, _ := resRaw["message"].(string)
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

	// 2. 获取容量和 VIP 信息
	// 如果有 kps，优先调用 App 接口获取 (能识别 88VIP 等细分等级)
	vipFetched := false
	if q.mparam["kps"] != "" {
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
				vipFetched = true
			}
		}
	}

	// 如果没有 kps 或者上面的 App 接口失败，降级使用 PC 端网页容量接口
	if !vipFetched {
		// 定义待探测的候选 URL 列表（优先尝试用户提供的最新 member 接口）
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

			// 解析探测
			dataNode, _ := capRaw["data"].(map[string]interface{})
			metadataNode, _ := capRaw["metadata"].(map[string]interface{})

			// 汇总可用的数据节点
			resNode := dataNode
			if resNode == nil {
				resNode = metadataNode
			}
			if resNode == nil {
				resNode = capRaw
			}

			if resNode != nil {
				// 1. 提取容量
				capInfo, _ := resNode["cap_info"].(map[string]interface{})
				if capInfo == nil {
					capInfo = resNode
				}

				total := float64(0)
				used := float64(0)

				// 兼容多种字段名：total/used (PC) vs cap_total/cap_used (User) vs total_capacity/use_capacity (Member)
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

				// 2. 提取 VIP 等级
				if mt, ok := resNode["member_type"]; ok {
					vipMap := map[string]string{
						"NORMAL":    "普通用户",
						"EXP_SVIP":  "88VIP",
						"SUPER_VIP": "SVIP",
						"Z_VIP":     "SVIP+",
					}

					switch v := mt.(type) {
					case string:
						if name, ok := vipMap[v]; ok {
							q.account.VipName = name
						} else {
							level, _ := strconv.Atoi(v)
							if level == 0 {
								q.account.VipName = "普通用户"
							} else if level == 1 {
								q.account.VipName = "VIP"
							} else if level == 2 {
								q.account.VipName = "SVIP"
							}
						}
					case float64:
						level := int(v)
						if level == 0 {
							q.account.VipName = "普通用户"
						} else if level == 1 {
							q.account.VipName = "VIP"
						} else if level == 2 {
							q.account.VipName = "SVIP"
						}
					}
				}

				if vipFetched {
					break // 成功获取，退出探测
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
	apiURL := BaseURL + "/1/clouddrive/file/sort"
	query := url.Values{}
	query.Set("pdir_fid", parentID)
	query.Set("_page", "1")
	query.Set("_size", "100")
	query.Set("_sort", "file_type:asc,updated_at:desc")

	resp, err := q.doRequest(ctx, "GET", apiURL, query, nil, false)
	if err != nil {
		return nil, err
	}

	var res struct {
		Code interface{} `json:"code"`
		Data struct {
			List []struct {
				Fid      string `json:"fid"`
				FileName string `json:"file_name"`
				Dir      bool   `json:"dir"`
				Size     int64  `json:"size"`
				UpdateAt int64  `json:"updated_at"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}

	codeStr := fmt.Sprintf("%v", res.Code)
	if codeStr != "0" && codeStr != "0.0" {
		return nil, fmt.Errorf("Quark API error: %v", res.Code)
	}

	var files []core.FileInfo
	for _, item := range res.Data.List {
		files = append(files, core.FileInfo{
			ID:        item.Fid,
			Name:      item.FileName,
			IsFolder:  item.Dir,
			Size:      item.Size,
			UpdatedAt: time.Unix(item.UpdateAt/1000, 0).Format("2006-01-02 15:04:05"),
		})
	}
	return files, nil
}

func (q *Quark) CreateFolder(ctx context.Context, name, parentID string) (string, error) {
	if parentID == "" || parentID == "/" {
		parentID = "0"
	}
	apiURL := BaseURL + "/1/clouddrive/file"
	body := map[string]interface{}{
		"pdir_fid":  parentID,
		"file_name": name,
		"dir_path":  "",
	}
	jsonBody, _ := json.Marshal(body)
	resp, err := q.doRequest(ctx, "POST", apiURL, nil, strings.NewReader(string(jsonBody)), false)
	if err != nil {
		return "", err
	}

	var res struct {
		Code interface{} `json:"code"`
		Data struct {
			Fid string `json:"fid"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return "", err
	}

	codeStr := fmt.Sprintf("%v", res.Code)
	if codeStr != "0" && codeStr != "0.0" {
		return "", fmt.Errorf("Quark API error: %v", res.Code)
	}
	return res.Data.Fid, nil
}

func (q *Quark) DeleteFile(ctx context.Context, fileID string) error {
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

func (q *Quark) SaveLink(ctx context.Context, shareURL, extractCode, targetPath string) error {
	// 1. 提取 pwd_id
	reID := regexp.MustCompile(`/s/(\w+)`)
	match := reID.FindStringSubmatch(shareURL)
	if len(match) < 2 {
		return fmt.Errorf("invalid quark share url")
	}
	pwdID := match[1]

	// 2. 获取 stoken
	tokenURL := BaseURL + "/1/clouddrive/share/sharepage/token"
	tokenBody := map[string]interface{}{
		"pwd_id":   pwdID,
		"passcode": extractCode,
	}
	jsonToken, _ := json.Marshal(tokenBody)
	resp, err := q.doRequest(ctx, "POST", tokenURL, nil, strings.NewReader(string(jsonToken)), false)
	if err != nil {
		return err
	}
	var tokenRes struct {
		Code interface{} `json:"code"`
		Data struct {
			Stoken string `json:"stoken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &tokenRes); err != nil {
		return err
	}

	codeStr := fmt.Sprintf("%v", tokenRes.Code)
	if codeStr != "0" && codeStr != "0.0" {
		return fmt.Errorf("Quark token error: %v", tokenRes.Code)
	}
	stoken := tokenRes.Data.Stoken

	// 3. 获取详情
	detailURL := BaseURL + "/1/clouddrive/share/sharepage/detail"
	detailQuery := url.Values{}
	detailQuery.Set("pwd_id", pwdID)
	detailQuery.Set("stoken", stoken)
	detailQuery.Set("pdir_fid", "0")
	resp, err = q.doRequest(ctx, "GET", detailURL, detailQuery, nil, false)
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

	// 4. 准备目标目录
	targetID, err := q.prepareTargetPath(ctx, targetPath)
	if err != nil {
		return err
	}

	// 5. 执行转存
	var fids []string
	var tokens []string
	for _, item := range detailRes.Data.List {
		fids = append(fids, item.Fid)
		tokens = append(tokens, item.ShareFidToken)
	}

	saveURL := BaseURL + "/1/clouddrive/share/sharepage/save"
	saveBody := map[string]interface{}{
		"fid_list":       fids,
		"fid_token_list": tokens,
		"to_pdir_fid":    targetID,
		"pwd_id":         pwdID,
		"stoken":         stoken,
		"pdir_fid":       "0",
		"scene":          "link",
	}
	jsonSave, _ := json.Marshal(saveBody)
	_, err = q.doRequest(ctx, "POST", saveURL, nil, strings.NewReader(string(jsonSave)), false)
	return err
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
			newID, err := q.CreateFolder(ctx, part, currentID)
			if err != nil {
				return "", err
			}
			currentID = newID
		}
	}
	return currentID, nil
}
