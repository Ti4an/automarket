// Lógica del panel: funciones del vendedor o del administrador según el rol.

const token = sessionStorage.getItem("token");
const role = sessionStorage.getItem("role");

// Sin sesión no hay panel: volver a la página principal.
if (!token) location.href = "/";

// Escapa texto para insertarlo de forma segura en HTML.
function esc(text) {
  const div = document.createElement("div");
  div.textContent = String(text);
  return div.innerHTML;
}

function money(value) {
  return "$" + Number(value).toLocaleString("es-CO");
}

// Llama a la API con el token; si la sesión expiró, vuelve al inicio.
async function api(method, url, body) {
  const res = await fetch(url, {
    method,
    headers: { Authorization: "Bearer " + token },
    body,
  });
  if (res.status === 401) {
    sessionStorage.clear();
    location.href = "/";
    return null;
  }
  const data = res.status === 204 ? null : await res.json();
  if (!res.ok) throw new Error(data.error || "Error");
  return data;
}

// Celdas comunes de una fila de vehículo.
function vehicleCells(v) {
  return `<td><img class="thumb" src="${esc(v.image_url)}" alt=""></td>
          <td>${esc(v.brand)} ${esc(v.model)} (${esc(v.year)})</td>
          <td>${money(v.price)}</td>`;
}

function statusBadge(status) {
  return `<span class="status status-${esc(status)}">${esc(status)}</span>`;
}

// ---------- Vendedor ----------
async function loadMyVehicles() {
  const vehicles = await api("GET", "/my/vehicles");
  document.getElementById("my-vehicles").innerHTML = vehicles.map((v) => `
    <tr>${vehicleCells(v)}<td>${statusBadge(v.status)}</td>
      <td>${v.status === "publicada"
        ? `<button class="btn btn-small" data-sold="${esc(v.id)}">Marcar vendido</button>` : ""}</td>
    </tr>`).join("") || `<tr><td colspan="5" class="muted">Sin publicaciones.</td></tr>`;
}

function initSeller() {
  document.getElementById("seller").classList.remove("hidden");

  // Envía el formulario con la imagen como multipart/form-data.
  document.getElementById("form-publish").onsubmit = async (e) => {
    e.preventDefault();
    const msg = document.getElementById("publish-msg");
    try {
      await api("POST", "/vehicles", new FormData(e.target));
      e.target.reset();
      msg.textContent = "Publicación enviada. Queda pendiente de aprobación.";
      loadMyVehicles();
    } catch (err) {
      msg.textContent = err.message;
    }
  };

  document.getElementById("my-vehicles").onclick = async (e) => {
    const id = e.target.dataset.sold;
    if (!id) return;
    try {
      await api("PATCH", `/vehicles/${encodeURIComponent(id)}/sold`);
      loadMyVehicles();
    } catch (err) {
      alert(err.message);
    }
  };

  loadMyVehicles();
}

// ---------- Administrador ----------
async function loadAdmin() {
  const vehicles = await api("GET", "/admin/vehicles");
  const users = await api("GET", "/admin/users");
  const pending = vehicles.filter((v) => v.status === "pendiente");

  document.getElementById("pending").innerHTML = pending.map((v) => `
    <tr>${vehicleCells(v)}<td class="nowrap">
      <button class="btn btn-small" data-approve="${esc(v.id)}">Aprobar</button>
      <button class="btn btn-small btn-danger" data-delete="${esc(v.id)}">Eliminar</button>
    </td></tr>`).join("") || `<tr><td colspan="4" class="muted">No hay pendientes.</td></tr>`;

  document.getElementById("all-vehicles").innerHTML = vehicles.map((v) => `
    <tr>${vehicleCells(v)}<td>${statusBadge(v.status)}</td><td>
      <button class="btn btn-small btn-danger" data-delete="${esc(v.id)}">Eliminar</button>
    </td></tr>`).join("") || `<tr><td colspan="5" class="muted">Sin publicaciones.</td></tr>`;

  document.getElementById("users").innerHTML = users.map((u) => `
    <tr><td>${esc(u.name)}</td><td>${esc(u.email)}</td><td>${esc(u.phone)}</td><td>
      <button class="btn btn-small btn-danger" data-user="${esc(u.id)}">Eliminar usuario</button>
    </td></tr>`).join("") || `<tr><td colspan="4" class="muted">No hay vendedores.</td></tr>`;
}

function initAdmin() {
  document.getElementById("admin").classList.remove("hidden");

  // Un solo manejador para todos los botones del panel admin.
  document.getElementById("admin").onclick = async (e) => {
    const { approve, delete: del, user } = e.target.dataset;
    try {
      if (approve) {
        await api("PATCH", `/vehicles/${encodeURIComponent(approve)}/approve`);
      } else if (del && confirm("¿Eliminar esta publicación?")) {
        await api("DELETE", `/vehicles/${encodeURIComponent(del)}`);
      } else if (user && confirm("¿Eliminar este usuario y sus publicaciones?")) {
        await api("DELETE", `/users/${encodeURIComponent(user)}`);
      } else {
        return;
      }
      loadAdmin();
    } catch (err) {
      alert(err.message);
    }
  };

  loadAdmin();
}

// ---------- Inicio ----------
document.getElementById("who").textContent = `${sessionStorage.getItem("name") || ""} · ${role}`;
document.getElementById("btn-logout").onclick = () => {
  sessionStorage.clear();
  location.href = "/";
};

if (role === "vendedor") initSeller();
else if (role === "administrador") initAdmin();