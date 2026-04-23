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
		respBody = `{"code": 0, "data": {"nickname": "E2E夸克用户"}}`
	} else if strings.Contains(url, "pan.quark.cn/1/clouddrive/member") || strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/capacity") {
		respBody = `{"code": 0, "data": {"total_capacity": 1099511627776, "use_capacity": 549755813888, "member_type": "SVIP"}}`
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/share/sharepage/detail") {
		// 模拟返回文件列表
		respBody = `{"code": 0, "data": {"list": [{"fid": "file1", "file_name": "[2024.04.20] E2E测试电影.mp4", "size": 1024, "updated_at": 1612345678000, "dir": false, "share_fid_token": "mock_token_1"}, {"fid": "file2", "file_name": "readme.txt", "size": 100, "updated_at": 1612345679000, "dir": false, "share_fid_token": "mock_token_2"}]}}`
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/share/sharepage/token") {
		respBody = `{"code": 0, "data": {"stoken": "mock_stoken"}}`
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/file/sort") {
		// 目标目录文件列表 (预检)
		respBody = `{"code": 0, "data": {"list": []}}`
	} else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/file") && req.Method == "POST" {
		// 创建目录
		respBody = `{"code": 0, "data": {"fid": "mock_dir_123"}}`
	}

	// 2. 模拟 139 相关接口
	if strings.Contains(url, "user-njs.yun.139.com/user/getUser") {
		respBody = `{"code": "0000", "success": true, "data": {"auditNickName": "E2E移动云盘用户", "userName": "E2E移动云盘用户", "userDomainId": "mock_domain", "loginName": "13800000000"}}`
	} else if strings.Contains(url, "user-njs.yun.139.com/user/disk/getPersonalDiskInfo") || strings.Contains(url, "user-njs.yun.139.com/user/disk/getFamilyDiskInfo") {
		respBody = `{"code": "0", "success": true, "data": {"diskSize": "1048576", "freeDiskSize": "524288"}}`
	} else if strings.Contains(url, "yun.139.com/orchestration/group-rebuild/member/v1.0/queryUserBenefits") {
		respBody = `{"code": "0", "success": true, "data": {"userSubMemberList": [{"memberLvName": "黄金会员"}]}}`
	} else if strings.Contains(url, "share-kd-njs.yun.139.com/yun-share/richlifeApp/devapp/IOutLink/getOutLinkInfoV6") {
		respBody = `{"code": "0", "data": {"coLst": [{"coID": "f1", "coName": "[2024.04.20] E2E测试电影.mp4", "size": 1024, "udTime": "20240420120000"}, {"coID": "f2", "coName": "readme.txt", "size": 100, "udTime": "20240420120100"}], "caLst": []}}`
	} else if strings.Contains(url, "share-kd-njs.yun.139.com/yun-share/richlifeApp/devapp/IBatchOprTask/createOuterLinkBatchOprTask") {
		respBody = `{"success": true}`
	} else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/list") {
		respBody = `{"code": "0", "success": true, "data": {"items": []}}`
	} else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/update") {
		respBody = `{"code": "0", "success": true}`
	} else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/create") {
		respBody = `{"code": "0", "success": true, "data": {"fileId": "mock_dir_139", "fileName": "mock_dir"}}`
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
