const API_BASE = window.API_BASE || '/api';

export async function apiRequest(path, {method='GET', body, token, headers={}} = {}){
  const opts = {method, headers: {...headers}}
  if(body){
    opts.body = JSON.stringify(body)
    opts.headers['Content-Type'] = 'application/json'
  }
  if(token){ opts.headers['Authorization'] = 'Bearer ' + token }
  const res = await fetch(API_BASE + path, opts)
  const text = await res.text()
  try{ return JSON.parse(text) }catch(e){ return text }
}
