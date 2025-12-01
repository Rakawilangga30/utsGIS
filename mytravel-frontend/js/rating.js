// Convert angka ke HTML bintang
function renderStars(rating = 0) {
    rating = Number(rating);
    let html = "";

    for (let i = 1; i <= 5; i++) {
        html += `<span style="color:#fbbf24; font-size:20px;">${i <= rating ? "★" : "☆"}</span>`;
    }
    return html;
}
