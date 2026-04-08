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
