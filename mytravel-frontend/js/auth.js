function login() {
    fetch(API + "/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            email: document.getElementById("email").value,
            password: document.getElementById("password").value
        })
    })
    .then(res => res.json())
    .then(d => {
        if (d.error) {
            alert(d.error);
        } else {
            alert("Login berhasil!");
            window.location.href = "index.html";
        }
    });
}
