import request from './request'

export function getAccounts() {
  return request({
    url: '/accounts',
    method: 'get'
  })
}

export function createAccount(data) {
  return request({
    url: '/accounts',
    method: 'post',
    data
  })
}

export function updateAccount(id, data) {
  return request({
    url: `/accounts/${id}`,
    method: 'put',
    data
  })
}

export function deleteAccount(id) {
  return request({
    url: `/accounts/${id}`,
    method: 'delete'
  })
}

export function checkAccount(id) {
  return request({
    url: `/accounts/${id}/check`,
    method: 'post'
  })
}

export function getFolders(accountId, parentID = '', parentPath = '/') {
  return request({
    url: `/accounts/${accountId}/folders`,
    method: 'get',
    params: { parent_id: parentID, parent_path: parentPath }
  })
}

export function createFolder(accountId, parentID, parentPath, name) {
  return request({
    url: `/accounts/${accountId}/folders`,
    method: 'post',
    data: { parent_id: parentID, parent_path: parentPath, name }
  })
}
