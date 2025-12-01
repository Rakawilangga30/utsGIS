// ==========================
// LOAD SEMUA TEMPAT WISATA
// ==========================
(async function(){
    const helper = window.apiRequest || (async (p,o)=>{ const r = await fetch((window.API||'')+p,{ credentials:'include', ...o }); return r.json() })
    const list = await helper('/api/admin/places')
    let html = ''
    ;(list||[]).forEach(p=>{
        html += `
            <div class="card">
                <h3>${p.name}</h3>
                <p>${p.category}</p>
                <small>Pemilik: ${p.created_by}</small><br><br>
                <button onclick="deletePlace('${p._id}')">Hapus</button>
            </div>
        `
    })
    const el = document.getElementById('adminPlaces')
    if(el) el.innerHTML = html
})();


// ==========================
// DELETE TEMPAT
// ==========================
window.deletePlace = async function(id){
  if(!confirm('Yakin ingin menghapus?')) return
  const helper = window.apiRequest || (async (p,o)=>{ const r = await fetch((window.API||'')+p,{ credentials:'include', ...o }); return r.json() })
  const d = await helper('/api/admin/delete-place/' + id, { method: 'DELETE' })
  alert(d.message || 'Selesai')
  location.reload()
}



// ==========================
// LOAD SEMUA USER
// ==========================
(async function(){
    const helper = window.apiRequest || (async (p,o)=>{ const r = await fetch((window.API||'')+p,{ credentials:'include', ...o }); return r.json() })
    const list = await helper('/api/admin/users')
    let html = ''
    ;(list||[]).forEach(u=>{
        html += `
            <div class="card">
                <b>${u.name}</b><br>
                <small>${u.email}</small><br>
                <span>Role: ${u.role}</span>
            </div>
        `
    })
    const el = document.getElementById('adminUsers')
    if(el) el.innerHTML = html
})();
