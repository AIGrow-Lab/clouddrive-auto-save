import request from './request'

export function getTasks() {
  return request({
    url: '/tasks',
    method: 'get'
  })
}

export function createTask(data) {
  return request({
    url: '/tasks',
    method: 'post',
    data
  })
}

export function updateTask(id, data) {
  return request({
    url: `/tasks/${id}`,
    method: 'put',
    data
  })
}

export function deleteTask(id) {
  return request({
    url: `/tasks/${id}`,
    method: 'delete'
  })
}

export function runTask(id) {
  return request({
    url: `/tasks/${id}/run`,
    method: 'post'
  })
}

export function dismissTask(id) {
  return request({
    url: `/tasks/${id}/dismiss`,
    method: 'post'
  })
}

export function previewTask(data) {
  return request({
    url: '/tasks/preview',
    method: 'post',
    data
  })
}

export function parseShareLink(data) {
  return request({
    url: '/tasks/parse_share',
    method: 'post',
    data
  })
}

export function getScheduleSettings() {
  return request({
    url: '/settings/schedule',
    method: 'get'
  })
}

export function updateScheduleSettings(data) {
  return request({
    url: '/settings/schedule',
    method: 'post',
    data
  })
}
