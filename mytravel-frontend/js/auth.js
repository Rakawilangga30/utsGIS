// Use global apiRequest (defined in js/api.js) and include credentials for session cookie
// Login User
async function login() {
  const email = document.getElementById('email')?.value
  const password = document.getElementById('password')?.value
  if(!email || !password){ alert('Email dan Password wajib diisi'); return }

  const helper = window.apiRequest || (async (p, o)=>{ // fallback
    const res = await fetch((window.API||'') + p, { method: o.method||'GET', headers: {'Content-Type':'application/json'}, body: JSON.stringify(o.body||{}), credentials: 'include' })
    return res.json()
  })

  const data = await helper('/api/auth/login', { method: 'POST', body: { email, password } })
  if(data.error){ alert(data.error); return }
  alert(data.message || 'Login berhasil')
  window.location.href = 'dashboard.html'
}

// Register User
async function registerUser(){
  const name = document.getElementById('name')?.value
  const email = document.getElementById('email')?.value
  const password = document.getElementById('password')?.value
  if(!name || !email || !password){ alert('Semua data wajib diisi'); return }

  const helper = window.apiRequest || (async (p, o)=>{ const res = await fetch((window.API||'') + p, { method: o.method||'GET', headers: {'Content-Type':'application/json'}, body: JSON.stringify(o.body||{}), credentials: 'include' }); return res.json() })
  const data = await helper('/api/auth/register', { method: 'POST', body: { name, email, password } })
  if(data.error){ alert(data.error); return }
  alert(data.message || 'Registrasi berhasil')
  window.location.href = 'login.html'
}

// Logout helper
async function logout(){
  const helper = window.apiRequest || (async (p,o)=>{ const res = await fetch((window.API||'')+p,{method:o.method||'GET',credentials:'include'}); return res.json() })
  await helper('/api/auth/logout', { method: 'POST' })
  window.location.href = 'login.html'
}

// Expose
window.login = login
window.registerUser = registerUser
window.logout = logout