fetch(API + "/api/places")
    .then(res => res.json())
    .then(data => {
        let html = "";
        data.forEach(p => {
            html += `
            <div class="card">
                <img class="place-img" src="${API}/api/photo/${p.photo_id}" />
                <h2>${p.name}</h2>
                <p>${p.description}</p>
                <a class="btn" href="detail.html?id=${p._id}">Detail</a>
            </div>
            `;
        });
        document.getElementById("places").innerHTML = html;
    });
