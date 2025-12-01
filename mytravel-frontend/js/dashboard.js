// ================================
// 1. Load tempat wisata milik user
// ================================
fetch(API + "/api/my-places")
    .then(res => res.json())
    .then(list => {
        let html = "";
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

        document.getElementById("myPlaces").innerHTML = html;
    });


// Hapus tempat
function deletePlace(id) {
    fetch(API + "/api/my-places/" + id, {
        method: "DELETE"
    })
    .then(r => r.json())
    .then(d => {
        alert(d.message);
        location.reload();
    });
}

function editPlace(id) {
    window.location.href = "add-place.html?edit=" + id;
}


// ================================
// 2. Load review milik user
// ================================
fetch(API + "/api/my-reviews")
    .then(res => res.json())
    .then(list => {
        let html = "";
        list.forEach(r => {
            html += `
                <div class="card">
                    <b>Tempat:</b> ${r.place_name}<br>
                    <b>Rating:</b> ${r.rating}<br>
                    <p>${r.comment}</p>
                    <button onclick="deleteReview('${r._id}')">Hapus Review</button>
                </div>
            `;
        });

        document.getElementById("myReviews").innerHTML = html;
    });

function deleteReview(id) {
    fetch(API + "/api/my-reviews/" + id, {
        method: "DELETE"
    })
    .then(r => r.json())
    .then(d => {
        alert(d.message);
        location.reload();
    });
}
