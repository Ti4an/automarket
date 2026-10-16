// Lógica de la página principal: catálogo, contacto, login y registro.

// Escapa texto para insertarlo de forma segura en HTML.
function esc(text) {
  const div = document.createElement("div");
  div.textContent = String(text);
  return div.innerHTML;
}

// Formatea el precio en pesos.
function money(value) {
  return "$" + Number(value).toLocaleString("es-CO");
}

// Envía JSON al servidor y devuelve la respuesta ya convertida.
async function postJSON(url, body) {
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || "Error");
  return data;
}

// Carga el catálogo público y pinta una tarjeta por vehículo.
async function loadCatalog() {
  const res = await fetch("/vehicles");
  const vehicles = await res.json();
  const grid = document.getElementById("catalog");
  document.getElementById("catalog-empty").classList.toggle("hidden", vehicles.length > 0);
  grid.innerHTML = vehicles.map((v) => `
    <article class="card">
      <img src="${esc(v.image_url)}" alt="${esc(v.brand)} ${esc(v.model)}">
      <div class="card-body">
        <h3>${esc(v.brand)} ${esc(v.model)}</h3>
        <p class="muted">Año ${esc(v.year)}</p>
        <p class="price">${money(v.price)}</p>
        <button class="btn" data-contact="${esc(v.id)}">Ver contacto</button>
      </div>
    </article>`).join("");
}

// Muestra el contacto del vendedor de un vehículo.
async function showContact(id) {
  const res = await fetch("/vehicles/" + encodeURIComponent(id));
  const data = await res.json();
  const box = document.getElementById("contact-info");
  box.innerHTML = res.ok
    ? `<p><strong>Nombre:</strong> ${esc(data.seller.name)}</p>
       <p><strong>Correo:</strong> ${esc(data.seller.email)}</p>
       <p><strong>Teléfono:</strong> ${esc(data.seller.phone)}</p>`
    : `<p class="msg">${esc(data.error)}</p>`;
  document.getElementById("dlg-contact").showModal();
}

// Abrir y cerrar los modales.
document.getElementById("btn-login").onclick = () => document.getElementById("dlg-login").showModal();
document.getElementById("btn-register").onclick = () => document.getElementById("dlg-register").showModal();
document.querySelectorAll("[data-close]").forEach((b) => {
  b.onclick = () => b.closest("dialog").close();
});

// Clic en "Ver contacto" dentro del catálogo.
document.getElementById("catalog").onclick = (e) => {
  const id = e.target.dataset.contact;
  if (id) showContact(id);
};

// Login: guarda token y rol en sessionStorage y va al panel.
document.getElementById("form-login").onsubmit = async (e) => {
  e.preventDefault();
  const form = Object.fromEntries(new FormData(e.target));
  try {
    const data = await postJSON("/login", form);
    sessionStorage.setItem("token", data.token);
    sessionStorage.setItem("role", data.role);
    sessionStorage.setItem("name", data.name);
    location.href = "/panel";
  } catch (err) {
    document.getElementById("login-msg").textContent = err.message;
  }
};

// Registro: crea la cuenta de vendedor y abre el login.
document.getElementById("form-register").onsubmit = async (e) => {
  e.preventDefault();
  const form = Object.fromEntries(new FormData(e.target));
  try {
    await postJSON("/register", form);
    e.target.reset();
    document.getElementById("dlg-register").close();
    document.getElementById("login-msg").textContent = "Cuenta creada. Ahora inicia sesión.";
    document.getElementById("dlg-login").showModal();
  } catch (err) {
    document.getElementById("register-msg").textContent = err.message;
  }
};

loadCatalog();