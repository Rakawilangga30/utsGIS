const url = new URLSearchParams(window.location.search);
const id = url.get('id');

(async function(){
    const helper = window.apiRequest || (async (p,o)=>{ const r = await fetch((window.API||'')+p,{ credentials:'include', ...o }); return r.json() })
    const places = await helper('/api/places')
    const place = (places||[]).find(p=>p._id === id)
    if(!place) return
    const el = document.getElementById('detail') || document.getElementById('place-detail')
    if(!el) return
    el.innerHTML = `
        <img class="place-img" src="${(window.API||'')}/api/photo/${place.photo_id}">
        <h2>${place.name}</h2>
        <p>${place.description}</p>
    `
})();

async function loadReviews(){
    const helper = window.apiRequest || (async (p,o)=>{ const r = await fetch((window.API||'')+p,{ credentials:'include', ...o }); return r.json() })
    const list = await helper('/api/reviews?place_id=' + id)
    const el = document.getElementById('reviews')
    if(!el) return
    let html = ''
    ;(list||[]).forEach(r=>{
        html += `
            <div class="card">
                <b>Rating:</b> ${r.rating}<br>
                <p>${r.comment}</p>
            </div>
        `
    })
    el.innerHTML = html
}
loadReviews();

async function addReview(){
    const rating = Number(document.getElementById('rating')?.value)
    const comment = document.getElementById('comment')?.value
    const helper = window.apiRequest || (async (p,o)=>{ const r = await fetch((window.API||'')+p,{ credentials:'include', ...o }); return r.json() })
    const d = await helper('/api/reviews', { method: 'POST', body: { place_id: id, rating, comment } })
    if(d.error) alert(d.error)
    else loadReviews()
}
