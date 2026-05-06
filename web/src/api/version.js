import request from './request'

export function getVersion() {
  return request.get('/version')
}
