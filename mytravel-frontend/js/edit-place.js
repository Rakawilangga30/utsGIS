const url = new URLSearchParams(window.location.search);
const placeId = url.get("id");

// =======================
// LOAD DATA TEMPAT LAMA
// =======================
fetch(API + "/api/places")
    .then(res => res.json())
    .then(data => {

        const place = data.find(p => p._id === placeId);
        if (!place) {
            alert("Data tidak ditemukan");
            return;
        }

        document.getElementById("name").value = place.name;
        document.getElementById("category").value = place.category;
        document.getElementById("description").value = place.description;
        document.getElementById("address").value = place.address;
        document.getElementById("lat").value = place.lat;
        document.getElementById("lng").value = place.lng;

        initMap(place.lat, place.lng); // load map
    });


// =======================
// MAP LEAFLET
// =======================
let map, marker;

function initMap(lat, lng) {
  
    map = L.map('map').setView([lat, lng], 14);

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 18,
    }).addTo(map);

    marker = L.marker([lat, lng]).addTo(map);

    // ketika map diklik → pindah marker
    map.on("click", (e) => {
        const { lat, lng } = e.latlng;

        marker.setLatLng([lat, lng]);

        document.getElementById("lat").value = lat;
        document.getElementById("lng").value = lng;
    });
}


// =======================
// SUBMIT FORM UPDATE
// =======================
document.getElementById("editForm").onsubmit = function (e) {
    e.preventDefault();

    const form = new FormData();

    form.append("name", document.getElementById("name").value);
    form.append("category", document.getElementById("category").value);
    form.append("description", document.getElementById("description").value);
    form.append("address", document.getElementById("address").value);
    form.append("lat", document.getElementById("lat").value);
    form.append("lng", document.getElementById("lng").value);

    const photoFile = document.getElementById("photo").files[0];
    if (photoFile) {
        form.append("photo", photoFile);
    }

    fetch(API + "/api/places/" + placeId, {
        method: "PUT",
        body: form
    })
        .then(res => res.json())
        .then(d => {
            if (d.error) {
                alert(d.error);
            } else {
                alert("Tempat berhasil diperbarui");
                window.location.href = "dashboard.html";
            }
        });
};
