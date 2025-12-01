// Simple auth helper that uses apiRequest if available
function saveToken(token){ try{ localStorage.setItem('mt_token', token) }catch(e){} }

async function login(){
    const email = document.getElementById('email')?.value
    const password = document.getElementById('password')?.value
    if(!email || !password){ alert('Email & password wajib'); return }

    try{
        const res = (window.apiRequest)
          ? await window.apiRequest('/auth/login', {method:'POST', body:{email,password}})
          : await fetch(window.API_BASE + '/auth/login', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({email,password})}).then(r=>r.json())

        if(res.error){ alert(res.error); return }
        // If backend returns token store it, otherwise store a minimal flag
        if(res.token) saveToken(res.token)
        alert(res.message || 'Login berhasil')
        window.location.href = 'index.html'
    }catch(err){
        console.error(err)
        alert('Login gagal, cek console untuk detil')
    }
}

async function registerUser(){
    const name = document.getElementById('name')?.value
    const email = document.getElementById('email')?.value
    const password = document.getElementById('password')?.value
    if(!email || !password){ alert('Email & password wajib'); return }

    try{
        const res = (window.apiRequest)
          ? await window.apiRequest('/auth/register', {method:'POST', body:{name,email,password}})
          : await fetch(window.API_BASE + '/auth/register', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name,email,password})}).then(r=>r.json())

        if(res.error){ alert(res.error); return }
        alert(res.message || 'Registrasi berhasil')
        window.location.href = 'login.html'
    }catch(err){ console.error(err); alert('Registrasi gagal') }
}

// expose to global for inline onclick usages
window.login = login
window.registerUser = registerUser

export { login, registerUser }
