(async function(){
    const helper = window.apiRequest || (async (p,o)=>{ const r = await fetch((window.API||'')+p, { credentials: 'include', ...o }); return r.json() })
    const data = await helper('/api/places')
    let html = ''
    ;(data||[]).forEach(p=>{
        html += `
            <div class="card">
                <img class="place-img" src="${(window.API||'')}/api/photo/${p.photo_id}" />
                <h2>${p.name}</h2>
                <p>${p.description || ''}</p>
                <a class="btn" href="detail.html?id=${p._id}">Detail</a>
            </div>
        `
    })
    const el = document.getElementById('places')
    if(el) el.innerHTML = html
})()
