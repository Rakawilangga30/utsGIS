let places = [];
let reviews = [];
let userLocation = null;

// Fungsi dipanggil dari HTML
function initRecommendation() {
    fetchPlaces();
}

// =============== FETCH PLACES ===============
function fetchPlaces() {
    fetch(API + "/api/places")
        .then(res => res.json())
        .then(data => {
            places = data;
            fetchReviews();
        });
}

function fetchReviews() {
    fetch(API + "/api/reviews")
        .then(res => res.json())
        .then(data => {
            reviews = data;

            // hitung rating
            places.forEach(p => {
                let rev = reviews.filter(r => r.place_id === p._id);
                p.reviewCount = rev.length;
                p.rating = rev.length ? average(rev.map(x => Number(x.rating))) : 0;
            });

            showRecommended();
        });
}

function average(arr) {
    return arr.reduce((a, b) => a + b, 0) / arr.length;
}


// =============== HITUNG JARAK (Haversine) ===============
function distance(lat1, lon1, lat2, lon2) {
    let R = 6371; // km
    let dLat = (lat2 - lat1) * Math.PI / 180;
    let dLon = (lon2 - lon1) * Math.PI / 180;

    let a =
        0.5 - Math.cos(dLat) / 2 +
        Math.cos(lat1 * Math.PI / 180) *
        Math.cos(lat2 * Math.PI / 180) *
        (1 - Math.cos(dLon)) / 2;

    return R * 2 * Math.asin(Math.sqrt(a));
}



// =============== REKOMENDASI ===============
function showRecommended() {

    let scored = places.map(p => {

        // default jarak (jika user tidak aktifkan lokasi)
        let dist = 999;

        if (userLocation) {
            dist = distance(userLocation.lat, userLocation.lng, p.lat, p.lng);
        }

        // Skor kombinasi:
        // rating * 2 + jumlah review + skor jarak
        let score = (p.rating * 2) + p.reviewCount + (100 - dist);

        return { ...p, dist, score };
    });

    // ranking terbaik
    scored.sort((a, b) => b.score - a.score);

    // ambil top 6
    scored = scored.slice(0, 6);

    let html = "";
    scored.forEach(p => {
        html += `
        <div class="card">

            <img src="${API}/api/photo/${p.photo_id}" class="place-img">

            <h2>${p.name}</h2>

            <div>
                ${renderStars(p.rating)}
                <small>(${p.reviewCount} review)</small>
            </div>

            <p>${p.description}</p>

            ${userLocation ? `<b>Jarak: ${p.dist.toFixed(1)} km</b><br><br>` : ""}

            <a class="btn" href="detail.html?id=${p._id}">Lihat Detail</a>

        </div>
        `;
    });

    document.getElementById("recom").innerHTML = html;
}



// =============== RENDER STAR UI ===============
function renderStars(rating) {
    let html = "";
    let full = Math.floor(rating);
    let half = rating - full >= 0.5;

    // full star
    for (let i = 0; i < full; i++) html += "⭐";

    // half star
    if (half) html += "⭐️";

    // empty star
    for (let i = html.length; i < 5; i++) html += "☆";

    return `<span style="font-size:20px;">${html}</span>`;
}
