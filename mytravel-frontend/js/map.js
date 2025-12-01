let map = L.map('map').setView([-6.9, 107.6], 12);

L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 18,
}).addTo(map);

let allPlaces = [];
let markerLayers = {};

// ======================
// LOAD SEMUA TEMPAT
// ======================
fetch(API + "/api/places")
    .then(res => res.json())
    .then(data => {
        allPlaces = data;
        renderCategories();
        renderMarkers();
    });


// ======================
// RENDER KATEGORI OTOMATIS
// ======================
function renderCategories() {
    let cats = [...new Set(allPlaces.map(p => p.category))];

    let html = "";
    cats.forEach(cat => {
        html += `
        <label>
            <input type="checkbox" class="filter-cat" value="${cat}" checked>
            ${cat}
        </label><br>
        `;
    });

    document.getElementById("categories").innerHTML = html;

    document.querySelectorAll(".filter-cat").forEach(cb => {
        cb.addEventListener("change", renderMarkers);
    });
}


// ======================
// TAMBAH MARKER KE MAP
// ======================
function renderMarkers() {

    // hapus marker lama
    for (let key in markerLayers) {
        map.removeLayer(markerLayers[key]);
    }
    markerLayers = {};

    let activeCats = [...document.querySelectorAll(".filter-cat:checked")].map(x => x.value);

    allPlaces.forEach(p => {
        if (!activeCats.includes(p.category)) return;

        let marker = L.marker([p.lat, p.lng]).addTo(map);

        marker.bindPopup(`
            <div style="text-align:center;">
                <img src="${API}/api/photo/${p.photo_id}" width="120" style="border-radius:6px;"><br><br>
                <b>${p.name}</b><br>
                <small>${p.category}</small><br><br>
                <a class="btn" href="detail.html?id=${p._id}">Lihat Detail</a>
            </div>
        `);

        markerLayers[p._id] = marker;
    });
}

function markerIcon(color) {
    return new L.Icon({
        iconUrl: `https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-${color}.png`,
        shadowUrl: "https://cdnjs.cloudflare.com/ajax/libs/leaflet/0.7.7/images/marker-shadow.png",
        iconSize: [25, 41],
        iconAnchor: [12, 41],
    });
}

const catColors = {
    "Wisata": "blue",
    "Kuliner": "red",
    "Belanja": "green",
    "Taman": "orange",
};
