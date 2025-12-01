// ==========================
// LOAD SEMUA TEMPAT WISATA
// ==========================
fetch(API + "/api/admin/places")
    .then(res => res.json())
    .then(list => {
        let html = "";
        list.forEach(p => {
            html += `
            <div class="card">
                <h3>${p.name}</h3>
                <p>${p.category}</p>
                <small>Pemilik: ${p.created_by}</small><br><br>

                <button onclick="deletePlace('${p._id}')">Hapus</button>
            </div>
            `;
        });

        document.getElementById("adminPlaces").innerHTML = html;
    });


// ==========================
// DELETE TEMPAT
// ==========================
function deletePlace(id) {
    fetch(API + "/api/admin/delete-place/" + id, {
        method: "DELETE"
    })
    .then(res => res.json())
    .then(d => {
        alert(d.message);
        location.reload();
    });
}



// ==========================
// LOAD SEMUA USER
// ==========================
fetch(API + "/api/admin/users")
    .then(res => res.json())
    .then(list => {
        let html = "";
        list.forEach(u => {
            html += `
            <div class="card">
                <b>${u.name}</b><br>
                <small>${u.email}</small><br>
                <span>Role: ${u.role}</span>
            </div>
            `;
        });

        document.getElementById("adminUsers").innerHTML = html;
    });
