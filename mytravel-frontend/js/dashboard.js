// Pastikan API diambil dari global (script.js)
const BASE_URL = window.API || window.API_BASE;

// ================================
// 1. Load tempat wisata milik user
// ================================
fetch(BASE_URL + "/api/my-places", { credentials: 'include' }) // Tambah credentials
    .then(res => res.json())
    .then(list => {
        let html = "";
        
        // Cek jika list kosong atau error
        if (!Array.isArray(list)) list = [];

        list.forEach(p => {
            html += `
            <div class="card">
                <h3>${p.name}</h3>
                <p>${p.category}</p>
                <button onclick="editPlace('${p._id}')">Edit</button>
                <button onclick="deletePlace('${p._id}')">Hapus</button>
            </div>
            `;
        });
        document.getElementById("myPlaces").innerHTML = html || "<p>Belum ada tempat wisata.</p>";
    });


// Hapus tempat
window.deletePlace = function(id) {
    if(!confirm("Yakin ingin menghapus?")) return;

    fetch(BASE_URL + "/api/my-places/" + id, {
        method: "DELETE",
        credentials: 'include' // <--- WAJIB ADA
    })
    .then(r => r.json())
    .then(d => {
        alert(d.message || "Berhasil dihapus");
        location.reload();
    });
}

window.editPlace = function(id) {
    window.location.href = "edit-place.html?id=" + id; // Perbaikan parameter URL (sebelumnya ?edit=)
}


// ================================
// 2. Load review milik user
// ================================
fetch(BASE_URL + "/api/my-reviews", { credentials: 'include' })
    .then(res => res.json())
    .then(list => {
        let html = "";
        if (!Array.isArray(list)) list = [];

        list.forEach(r => {
            html += `
                <div class="card">
                    <b>Tempat:</b> ${r.place_name || 'Tanpa Nama'}<br>
                    <b>Rating:</b> ${r.rating}<br>
                    <p>${r.comment}</p>
                    <button onclick="deleteReview('${r._id}')">Hapus Review</button>
                </div>
            `;
        });
        document.getElementById("myReviews").innerHTML = html || "<p>Belum ada review.</p>";
    });

window.deleteReview = function(id) {
    if(!confirm("Hapus review ini?")) return;

    fetch(BASE_URL + "/api/my-reviews/" + id, {
        method: "DELETE",
        credentials: 'include'
    })
    .then(r => r.json())
    .then(d => {
        alert(d.message || "Review dihapus");
        location.reload();
    });
}