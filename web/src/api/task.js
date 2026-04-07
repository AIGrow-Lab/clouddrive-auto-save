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
