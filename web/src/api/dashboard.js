import request from './request'

export function getStats() {
  return request({
    url: '/dashboard/stats',
    method: 'get'
  })
}

export function clearLogsAPI() {
  return request({
    url: '/dashboard/logs/recent',
    method: 'delete'
  })
}
