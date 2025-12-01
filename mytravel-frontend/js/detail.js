const url = new URLSearchParams(window.location.search);
const id = url.get("id");

fetch(API + "/api/places")
    .then(res => res.json())
    .then(data => {
        const place = data.find(p => p._id === id);
        if (!place) return;

        document.getElementById("detail").innerHTML = `
            <img class="place-img" src="${API}/api/photo/${place.photo_id}">
            <h2>${place.name}</h2>
            <p>${place.description}</p>
        `;
    });

function loadReviews() {
    fetch(API + "/api/reviews?place_id=" + id)
        .then(res => res.json())
        .then(list => {
            let html = "";
            list.forEach(r => {
                html += `
                <div class="card">
                    <b>Rating:</b> ${r.rating}<br>
                    <p>${r.comment}</p>
                </div>
                `;
            });
            document.getElementById("reviews").innerHTML = html;
        });
}
loadReviews();

function addReview() {
    fetch(API + "/api/reviews", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            place_id: id,
            rating: Number(document.getElementById("rating").value),
            comment: document.getElementById("comment").value
        })
    })
    .then(r => r.json())
    .then(d => {
        if (d.error) alert(d.error);
        else loadReviews();
    });
}

document.getElementById("detail").innerHTML = `
    <img class="place-img" src="${API}/api/photo/${place.photo_id}">
    <h2>${place.name}</h2>
    <p>${place.description}</p>

    <div style="margin:10px 0;">
        <b>Rating:</b>
        ${renderStars(place.rating || 0)}
        (${place.rating?.toFixed(1) || 0})
    </div>
`;
