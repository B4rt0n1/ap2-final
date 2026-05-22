const state = {
  userId: "",
  carId: "",
  bookingId: "",
};

const responseLog = document.querySelector("#responseLog");
const carsList = document.querySelector("#carsList");
const bookingsList = document.querySelector("#bookingsList");
const userIdInput = document.querySelector("#userIdInput");
const carIdInput = document.querySelector("#carIdInput");
const bookingForm = document.querySelector("#bookingForm");

const today = new Date();
const start = new Date(today);
start.setDate(today.getDate() + 1);
const end = new Date(today);
end.setDate(today.getDate() + 4);
bookingForm.start_date.value = toDateInput(start);
bookingForm.end_date.value = toDateInput(end);

document.querySelector("#registerForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = formJSON(event.currentTarget);
  const response = await request("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  const user = response.user;
  if (user?.id) {
    setUser(user.id, `${user.first_name || payload.first_name} ${user.last_name || payload.last_name}`);
  }
});

document.querySelector("#carForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = formJSON(event.currentTarget);
  payload.year = Number(payload.year);
  payload.price_per_day = Number(payload.price_per_day);
  const response = await request("/api/cars", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  if (response.car?.id) {
    setCar(response.car);
  }
  await loadCars();
});

bookingForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = formJSON(event.currentTarget);
  payload.user_id = userIdInput.value.trim();
  if (!payload.user_id) {
    showError("Register a user or paste an active user ID.");
    return;
  }
  const response = await request("/api/bookings", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  if (response.booking?.id) {
    setBooking(response.booking);
    await loadBookings();
  }
});

document.querySelector("#confirmButton").addEventListener("click", () => changeBooking("confirm"));
document.querySelector("#cancelButton").addEventListener("click", () => changeBooking("cancel"));
document.querySelector("#refreshCarsButton").addEventListener("click", loadCars);
document.querySelector("#loadBookingsButton").addEventListener("click", loadBookings);
userIdInput.addEventListener("input", () => setUser(userIdInput.value.trim(), "Manual ID"));

async function changeBooking(action) {
  if (!state.bookingId) {
    showError("Create a booking first.");
    return;
  }
  const response = await request(`/api/bookings/${state.bookingId}/${action}`, { method: "POST" });
  if (response.booking) {
    setBooking(response.booking);
  }
  await loadBookings();
}

async function loadCars() {
  try {
    const response = await request("/api/cars");
    const cars = response.cars || [];
    if (!cars.length) {
      carsList.innerHTML = '<p class="empty">No cars yet. Add one in the Fleet form.</p>';
      return;
    }
    carsList.innerHTML = "";
    cars.forEach((car) => {
      const row = document.createElement("article");
      row.className = "car-row";
      row.innerHTML = `
        <strong>${escapeHTML(car.brand)} ${escapeHTML(car.model)}</strong>
        <span class="meta">${escapeHTML(car.category)} | ${car.year} | ${car.available ? "available" : "busy"}</span>
        <span class="price">$${Number(car.price_per_day || 0).toFixed(2)}/day</span>
        <button type="button">Select car</button>
      `;
      row.querySelector("button").addEventListener("click", () => setCar(car));
      carsList.append(row);
    });
  } catch (error) {
    carsList.innerHTML = `<p class="empty error">${escapeHTML(error.message)}</p>`;
  }
}

async function loadBookings() {
  const userId = userIdInput.value.trim();
  if (!userId) {
    showError("Choose a user before loading bookings.");
    return;
  }
  const response = await request(`/api/users/${encodeURIComponent(userId)}/bookings`);
  const bookings = response.bookings || [];
  if (!bookings.length) {
    bookingsList.innerHTML = '<p class="empty">No bookings for this renter.</p>';
    return;
  }
  bookingsList.innerHTML = "";
  bookings.forEach((booking) => {
    const row = document.createElement("article");
    row.className = "booking-row";
    row.innerHTML = `
      <strong>${escapeHTML(booking.status)} booking</strong>
      <span class="meta">${escapeHTML(booking.id)}</span>
      <span class="meta">Car ${escapeHTML(booking.car_id)}</span>
      <span class="price">$${Number(booking.total_price || 0).toFixed(2)}</span>
      <button type="button">Use booking</button>
    `;
    row.querySelector("button").addEventListener("click", () => setBooking(booking));
    bookingsList.append(row);
  });
}

async function request(url, options = {}) {
  const response = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  const text = await response.text();
  const body = text ? JSON.parse(text) : {};
  responseLog.textContent = JSON.stringify(body, null, 2);
  if (!response.ok) {
    throw new Error(body.error || body.message || `Request failed with ${response.status}`);
  }
  return body;
}

function setUser(id, label) {
  state.userId = id;
  userIdInput.value = id;
  document.querySelector("#activeUserLabel").textContent = id ? `${label}: ${shortID(id)}` : "Not selected";
}

function setCar(car) {
  state.carId = car.id;
  carIdInput.value = car.id;
  document.querySelector("#activeCarLabel").textContent = `${car.brand} ${car.model}: ${shortID(car.id)}`;
}

function setBooking(booking) {
  state.bookingId = booking.id;
  document.querySelector("#activeBookingLabel").textContent = `${booking.status}: ${shortID(booking.id)}`;
}

function showError(message) {
  responseLog.textContent = JSON.stringify({ error: message }, null, 2);
}

function formJSON(form) {
  return Object.fromEntries(new FormData(form).entries());
}

function shortID(value) {
  return value.length > 12 ? `${value.slice(0, 12)}...` : value;
}

function toDateInput(date) {
  return date.toISOString().slice(0, 10);
}

function escapeHTML(value = "") {
  const node = document.createElement("span");
  node.textContent = String(value);
  return node.innerHTML;
}

loadCars();
