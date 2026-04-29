package core

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

// 用于 E2E 测试的全局状态：记录哪些 FID 已被“转存成功”
var (
	savedFiles   = make(map[string]bool)
	savedFilesMu sync.Mutex
)

// ResetMockState 重置 Mock 状态
func ResetMockState() {
	savedFilesMu.Lock()
	savedFiles = make(map[string]bool)
	savedFilesMu.Unlock()
	slog.Info("E2E Mock State Reset")
}

// mockTransport 拦截并模拟网盘底层 API 请求
type mockTransport struct{}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	// 统一输出拦截日志，方便排错
	slog.Info("[HTTP Mock Intercepted]", "method", req.Method, "url", url)

	var respBody string
	var statusCode int = 200

	// 1. 模拟夸克相关接口
	if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/share/sharepage/save") {
		// 记录转存成功状态
		savedFilesMu.Lock()
		savedFiles["file1"] = true
		savedFiles["file2"] = true
		savedFilesMu.Unlock()
		respBody = `{"code": 0, "message": "ok", "data": {"task_id": "mock_task_123"}}`
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/task") {
		// 模拟任务成功
		respBody = `{"code": 0, "message": "ok", "data": {"status": 2, "message": "success"}}`
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/file/rename") {
		respBody = `{"code": 0, "message": "ok"}`
	} else if strings.Contains(url, "pan.quark.cn/account/info") {
		nickname := "E2E夸克用户"
		if strings.Contains(req.Header.Get("Cookie"), "mock_normal") {
			nickname = "E2E普通用户"
		} else if strings.Contains(req.Header.Get("Cookie"), "mock_overcap") {
			nickname = "E2E超容用户"
		} else if strings.Contains(req.Header.Get("Cookie"), "mock_svipplus") {
			nickname = "E2E超超级会员"
		}
		respBody = `{"code": 0, "data": {"nickname": "` + nickname + `"}}`
	} else if strings.Contains(url, "pan.quark.cn/1/clouddrive/member") || strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/capacity") {
		// 根据 Cookie 动态返回会员与容量
		memberType := "SUPER_VIP"
		totalCap := "1099511627776" // 1TB
		usedCap := "549755813888"   // 512GB

		if strings.Contains(req.Header.Get("Cookie"), "mock_normal") {
			memberType = "NORMAL"
			totalCap = "10737418240" // 10GB
			usedCap = "1073741824"   // 1GB
		} else if strings.Contains(req.Header.Get("Cookie"), "mock_overcap") {
			if strings.Contains(url, "getFamilyDiskInfo") {
				diskSize := "0"
				freeDiskSize := "0"
				respBody = `{"code": "0", "success": true, "data": {"diskSize": "` + diskSize + `", "freeDiskSize": "` + freeDiskSize + `"}}`
			} else {
				memberType = "SUPER_VIP"
				totalCap = "1099511627776" // 1TB
				usedCap = "2199023255552"  // 2TB (Over-capacity)
			}
		} else if strings.Contains(req.Header.Get("Cookie"), "mock_svipplus") {
			memberType = "Z_VIP"
			totalCap = "10995116277760" // 10TB
			usedCap = "1073741824"      // 1GB
		}

		if respBody == "" {
			respBody = `{"code": 0, "member_type": "` + memberType + `", "data": {"total_capacity": ` + totalCap + `, "used_capacity": ` + usedCap + `, "use_capacity": ` + usedCap + `, "member_type": "` + memberType + `"}}`
		}
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/share/sharepage/detail") {
	if strings.Contains(url, "mock_violation") {
		statusCode = 403
		respBody = `{"code": 41010, "message": "该分享文件涉及违规内容，已被官方屏蔽。"}`
	} else if strings.Contains(url, "mock_invalid") {
		statusCode = 404
		respBody = `{"code": 24001, "message": "该分享已失效，可能已被取消或删除。"}`
	} else if strings.Contains(url, "mock_empty") {
		respBody = `{"code": 0, "message": "ok", "data": {"list": []}}`
	} else if strings.Contains(url, "mock_quark_missing_code") {			respBody = `{"code": 41008, "message": "请先输入提取码"}`
		} else if strings.Contains(url, "mock_quark_wrong_code") {
			respBody = `{"code": 41007, "message": "提取码错误"}`
		} else {
			// 模拟返回文件列表
			respBody = `{"code": 0, "data": {"list": [{"fid": "file1", "file_name": "[2024.04.20] E2E测试电影.mp4", "size": 1024, "updated_at": 1612345678000, "dir": false, "share_fid_token": "mock_token_1"}, {"fid": "file2", "file_name": "readme.txt", "size": 100, "updated_at": 1612345679000, "dir": false, "share_fid_token": "mock_token_2"}]}}`
		}
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/share/sharepage/token") {
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		bodyStr := string(bodyBytes)

		if strings.Contains(bodyStr, "mock_violation") {
			statusCode = 403
			respBody = `{"code": 41010, "message": "该分享文件涉及违规内容，已被官方屏蔽。"}`
		} else if strings.Contains(bodyStr, "mock_invalid") {
			statusCode = 404
			respBody = `{"code": 24001, "message": "该分享已失效，可能已被取消或删除。"}`
		} else if strings.Contains(bodyStr, "mock_quark_missing_code") {
			respBody = `{"code": 41008, "message": "请先输入提取码"}`
		} else if strings.Contains(bodyStr, "mock_quark_wrong_code") {
			respBody = `{"code": 41007, "message": "提取码错误"}`
		} else {
			respBody = `{"code": 0, "data": {"stoken": "mock_stoken"}}`
		}
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/file/sort") {
		// 目标目录文件列表
		if strings.Contains(url, "pdir_fid=0") {
			respBody = `{"code": 0, "data": {"list": [{"fid": "quark_exist_dir", "file_name": "夸克已有目录", "dir": true, "size": 0, "updated_at": 1612345678000}]}}`
		} else if strings.Contains(url, "pdir_fid=quark_exist_dir") {
			respBody = `{"code": 0, "data": {"list": [{"fid": "quark_new_folder", "file_name": "quark_new_folder", "dir": true, "size": 0, "updated_at": 1612345679000}]}}`
		} else {
			savedFilesMu.Lock()
			if savedFiles["file1"] {
				respBody = `{"code": 0, "data": {"list": [{"fid": "file1", "file_name": "[2024.04.20] E2E测试电影.mp4", "size": 1024, "updated_at": 1612345678000, "dir": false}, {"fid": "file2", "file_name": "readme.txt", "size": 100, "updated_at": 1612345679000, "dir": false}]}}`
			} else {
				respBody = `{"code": 0, "data": {"list": []}}`
			}
			savedFilesMu.Unlock()
		}
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/file") && req.Method == "POST" {
		respBody = `{"code": 0, "data": {"fid": "quark_new_folder"}}`
	}

	// 2. 模拟 139 相关接口
	if strings.Contains(url, "user-njs.yun.139.com/user/getUser") {
		nickname := "E2E移动云盘用户"
		if strings.Contains(req.Header.Get("Authorization"), "mock_normal") {
			nickname = "E2E139普通用户"
		} else if strings.Contains(req.Header.Get("Authorization"), "mock_overcap") {
			nickname = "E2E139超容用户"
		} else if strings.Contains(req.Header.Get("Authorization"), "mock_silver") {
			nickname = "E2E139白银会员"
		}
		respBody = `{"code": "0000", "success": true, "data": {"auditNickName": "` + nickname + `", "userName": "` + nickname + `", "userDomainId": "mock_domain", "loginName": "13800000000"}}`
	} else if strings.Contains(url, "user-njs.yun.139.com/user/disk/getPersonalDiskInfo") || strings.Contains(url, "user-njs.yun.139.com/user/disk/getFamilyDiskInfo") {
		// 返回 MB 单位
		diskSize := "1048576"    // 1TB
		freeDiskSize := "524288" // 512GB

		if strings.Contains(req.Header.Get("Authorization"), "mock_normal") {
			diskSize = "20480"     // 20GB
			freeDiskSize = "10240" // 10GB
		} else if strings.Contains(req.Header.Get("Authorization"), "mock_overcap") {
			if strings.Contains(url, "getFamilyDiskInfo") {
				diskSize = "0"
				freeDiskSize = "0"
			} else {
				diskSize = "1048576"
				freeDiskSize = "-1048576"
			}
		}
		respBody = `{"code": "0", "success": true, "data": {"diskSize": "` + diskSize + `", "freeDiskSize": "` + freeDiskSize + `"}}`
	} else if strings.Contains(url, "yun.139.com/orchestration/group-rebuild/member/v1.0/queryUserBenefits") {
		vipName := "黄金会员"
		if strings.Contains(req.Header.Get("Authorization"), "mock_normal") {
			vipName = "普通用户"
		} else if strings.Contains(req.Header.Get("Authorization"), "mock_overcap") {
			vipName = "钻石会员"
		} else if strings.Contains(req.Header.Get("Authorization"), "mock_silver") {
			vipName = "白银会员"
		}
		respBody = `{"code": "0", "success": true, "data": {"userSubMemberList": [{"memberLvName": "` + vipName + `"}]}}`
	} else if strings.Contains(url, "share-kd-njs.yun.139.com/yun-share/richlifeApp/devapp/IOutLink/getOutLinkInfoV6") {
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		bodyStr := string(bodyBytes)

		if strings.Contains(bodyStr, "mock_invalid") {
			respBody = `{"code": "200000727", "message": "分享链接不存在或已被取消。"}`
		} else if strings.Contains(bodyStr, "mock_139_wrong_code") {
			respBody = `{"code": "9188", "message": "提取码校验失败"}`
		} else {
			// 提供完整的 path 以确保 ParseShare 和 SaveLink 标识符一致
			respBody = `{"code": "0", "data": {"coLst": [{"coID": "f1", "contentID": "f1", "parentCatalogID": "root", "path": "root/f1", "coName": "[2024.04.20] E2E测试电影.mp4", "size": 1024, "udTime": "20240420120000"}, {"coID": "f2", "contentID": "f2", "parentCatalogID": "root", "path": "root/f2", "coName": "readme.txt", "size": 100, "udTime": "20240420120100"}], "caLst": []}}`
		}
	} else if strings.Contains(url, "share-kd-njs.yun.139.com/yun-share/richlifeApp/devapp/IBatchOprTask/createOuterLinkBatchOprTask") {
		savedFilesMu.Lock()
		savedFiles["f1"] = true
		savedFiles["f2"] = true
		savedFilesMu.Unlock()
		respBody = `{"success": true}`
	} else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/list") {
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		bodyStr := string(bodyBytes)
		if strings.Contains(bodyStr, `"parentFileId":"root"`) || strings.Contains(bodyStr, `"parentFileId":"/"`) {
			respBody = `{"code": "0", "success": true, "data": {"items": [{"fileId": "139_exist_dir", "name": "139已有目录", "type": "folder", "category": "folder", "size": 0, "updatedAt": "2024-04-20 12:00:00"}]}}`
		} else if strings.Contains(bodyStr, `"parentFileId":"139_exist_dir"`) {
			respBody = `{"code": "0", "success": true, "data": {"items": [{"fileId": "139_new_folder", "name": "139_new_folder", "type": "folder", "category": "folder", "size": 0, "updatedAt": "2024-04-20 12:00:01"}]}}`
		} else {
			savedFilesMu.Lock()
			if savedFiles["f1"] {
				respBody = `{"code": "0", "success": true, "data": {"items": [{"fileId": "f1", "name": "[2024.04.20] E2E测试电影.mp4", "type": "file", "size": 1024, "updatedAt": "2024-04-20 12:00:00"}, {"fileId": "f2", "name": "readme.txt", "type": "file", "size": 100, "updatedAt": "2024-04-20 12:00:00"}]}}`
			} else {
				respBody = `{"code": "0", "success": true, "data": {"items": []}}`
			}
			savedFilesMu.Unlock()
		}
	} else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/update") {
		respBody = `{"code": "0", "success": true}`
	} else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/create") {
		respBody = `{"code": "0", "success": true, "data": {"fileId": "139_new_folder", "fileName": "139_new_folder"}}`
	}

	if respBody == "" {
		respBody = `{"code": 0, "success": true, "message": "unhandled mock route"}`
	}

	slog.Info("[HTTP Mock Response]", "url", url, "body", respBody)

	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(respBody)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

// SetupE2EHTTPMock 将全局的 HTTP 传输层替换为 Mock 拦截器
func SetupE2EHTTPMock() {
	HTTPTransport = &mockTransport{}
	slog.Info("E2E HTTP Transport Mock Enabled")
}
