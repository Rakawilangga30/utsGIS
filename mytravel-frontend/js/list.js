let allPlaces = [];

// =========================
// LOAD DATA
// =========================
fetch(API + "/api/places")
    .then(res => res.json())
    .then(data => {
        allPlaces = data;
        renderCategories();
        renderList();
    });


// =========================
// GENERATE FILTER KATEGORI
// =========================
function renderCategories() {
    let cats = [...new Set(allPlaces.map(p => p.category))];

    cats.forEach(cat => {
        let opt = document.createElement("option");
        opt.value = cat;
        opt.innerText = cat;
        document.getElementById("filterCat").appendChild(opt);
    });

    document.getElementById("filterCat").onchange = renderList;
    document.getElementById("sort").onchange = renderList;
}


// =========================
// RENDER LIST
// =========================
function renderList() {

    let cat = document.getElementById("filterCat").value;
    let sort = document.getElementById("sort").value;

    let filtered = allPlaces.filter(p => cat === "all" || p.category === cat);

    // Sorting
    if (sort === "newest") {
        filtered.sort((a, b) => b.created_at - a.created_at);
    }
    if (sort === "rating") {
        filtered.sort((a, b) => (b.rating || 0) - (a.rating || 0));
    }
    if (sort === "name") {
        filtered.sort((a, b) => a.name.localeCompare(b.name));
    }

    let html = "";
    filtered.forEach(p => {
        html += `
        <div class="card">
            <img src="${API}/api/photo/${p.photo_id}" class="place-img">

            <h2>${p.name}</h2>
            <p>${p.description}</p>

            <small>Kategori: ${p.category}</small><br>
            <small>Rating: ${p.rating || 0}</small><br><br>

            <a class="btn" href="detail.html?id=${p._id}">Lihat Detail</a>
        </div>
        `;
    });

    document.getElementById("list").innerHTML = html;
}
