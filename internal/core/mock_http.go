package core

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// mockTransport 拦截并模拟网盘底层 API 请求
type mockTransport struct{}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	// 统一输出拦截日志，方便排错
	slog.Info("[HTTP Mock Intercepted]", "method", req.Method, "url", url)

	var respBody string

	// 1. 模拟夸克相关接口
	if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/share/sharepage/save") {
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
			memberType = "SUPER_VIP"
			totalCap = "1099511627776" // 1TB
			usedCap = "2199023255552"  // 2TB (Over-capacity)
		} else if strings.Contains(req.Header.Get("Cookie"), "mock_svipplus") {
			memberType = "Z_VIP"
			totalCap = "10995116277760" // 10TB
			usedCap = "1073741824"      // 1GB
		}

		respBody = `{"code": 0, "member_type": "` + memberType + `", "data": {"total_capacity": ` + totalCap + `, "used_capacity": ` + usedCap + `, "use_capacity": ` + usedCap + `, "member_type": "` + memberType + `"}}`
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/share/sharepage/detail") {
		// 模拟返回文件列表
		respBody = `{"code": 0, "data": {"list": [{"fid": "file1", "file_name": "[2024.04.20] E2E测试电影.mp4", "size": 1024, "updated_at": 1612345678000, "dir": false, "share_fid_token": "mock_token_1"}, {"fid": "file2", "file_name": "readme.txt", "size": 100, "updated_at": 1612345679000, "dir": false, "share_fid_token": "mock_token_2"}]}}`
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/share/sharepage/token") {
		respBody = `{"code": 0, "data": {"stoken": "mock_stoken"}}`
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/file/sort") {
		// 目标目录文件列表 (预检)
		if strings.Contains(url, "pdir_fid=0") {
			respBody = `{"code": 0, "data": {"list": [{"fid": "quark_exist_dir", "file_name": "夸克已有目录", "dir": true, "size": 0, "updated_at": 1612345678000}]}}`
		} else if strings.Contains(url, "pdir_fid=quark_exist_dir") {
			respBody = `{"code": 0, "data": {"list": [{"fid": "quark_new_folder", "file_name": "quark_new_folder", "dir": true, "size": 0, "updated_at": 1612345679000}]}}`
		} else {
			respBody = `{"code": 0, "data": {"list": []}}`
		}
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/file") && req.Method == "POST" {
		// 创建目录
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
		diskSize := "1048576"   // 1TB (1024 * 1024 MB)
		freeDiskSize := "524288" // 512GB (512 * 1024 MB)

		if strings.Contains(req.Header.Get("Authorization"), "mock_normal") {
			diskSize = "20480"     // 20GB
			freeDiskSize = "10240" // 10GB
		} else if strings.Contains(req.Header.Get("Authorization"), "mock_overcap") {
			if strings.Contains(url, "getFamilyDiskInfo") {
				diskSize = "0"
				freeDiskSize = "0"
			} else {
				diskSize = "1048576"      // 1TB
				freeDiskSize = "-1048576" // -1TB -> Used = 2TB. Excess = 1TB.
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
		respBody = `{"code": "0", "data": {"coLst": [{"coID": "f1", "coName": "[2024.04.20] E2E测试电影.mp4", "size": 1024, "udTime": "20240420120000"}, {"coID": "f2", "coName": "readme.txt", "size": 100, "udTime": "20240420120100"}], "caLst": []}}`
	} else if strings.Contains(url, "share-kd-njs.yun.139.com/yun-share/richlifeApp/devapp/IBatchOprTask/createOuterLinkBatchOprTask") {
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
			respBody = `{"code": "0", "success": true, "data": {"items": []}}`
		}
	} else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/update") {
		respBody = `{"code": "0", "success": true}`
	} else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/create") {
		respBody = `{"code": "0", "success": true, "data": {"fileId": "139_new_folder", "fileName": "139_new_folder"}}`
	}

	if respBody == "" {
		respBody = `{"code": 0, "success": true, "message": "unhandled mock route"}`
	}

	return &http.Response{
		StatusCode: 200,
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
