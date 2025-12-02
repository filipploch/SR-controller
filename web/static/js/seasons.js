//let seasons = [];

// Ładowanie sezonów
async function loadSeasons() {
    try {
        const response = await fetch('/api/seasons');
        seasons = await response.json();
        renderSeasons();
    } catch (error) {
        console.error('Błąd ładowania sezonów:', error);
    }
}

// Renderowanie tabeli
function renderSeasons() {
    const tbody = document.getElementById('seasonsTableBody');

    if (seasons.length === 0) {
        tbody.innerHTML = `
            <tr>
                <td colspan="4">
                    <div class="empty-state">
                        <div class="empty-state-icon">📺</div>
                        <div>Brak sezonów. Utwórz pierwszy sezon.</div>
                    </div>
                </td>
            </tr>
        `;
        return;
    }

    tbody.innerHTML = seasons.map(season => `
        <tr>
            <td><strong>${season.number}</strong></td>
            <td>${season.description || '<em style="color: #666;">Brak opisu</em>'}</td>
            <td>
                ${season.is_current ? '<span class="badge badge-success">Aktualny</span>' : '<span class="badge badge-secondary">-</span>'}
            </td>
            <td>
                <div class="table-actions">
                    ${!season.is_current ? `<button class="btn btn-success btn-small" onclick="setCurrentSeason(${season.id})">Aktywuj</button>` : ''}
                    <button class="btn btn-primary btn-small" onclick="openEditSeasonModal(${season.id})">Edytuj</button>
                    <button class="btn btn-danger btn-small" onclick="deleteSeason(${season.id})">Usuń</button>
                </div>
            </td>
        </tr>
    `).join('');
}

// Otwórz modal tworzenia
async function openCreateModal() {
    document.getElementById('modalSeasonTitle').textContent = 'Nowy Sezon';
    document.getElementById('seasonForm').reset();
    document.getElementById('seasonId').value = '';

    // Pobierz następny numer sezonu
    try {
        const response = await fetch('/api/seasons/next-number');
        const data = await response.json();
        document.getElementById('seasonNumber').value = data.next_number;
    } catch (error) {
        console.error('Błąd pobierania następnego numeru:', error);
        document.getElementById('seasonNumber').value = 1;
    }

    document.getElementById('seasonModal').classList.add('active');
}

// Otwórz modal edycji
function openEditSeasonModal(id) {
    const season = seasons.find(s => s.id === id);
    if (!season) return;

    document.getElementById('modalSeasonTitle').textContent = 'Edycja Sezonu';
    document.getElementById('seasonId').value = season.id;
    document.getElementById('seasonNumber').value = season.number;
    document.getElementById('seasonDescription').value = season.description || '';
    document.getElementById('seasonIsCurrent').checked = season.is_current;
    document.getElementById('seasonModal').classList.add('active');
}

// Zamknij modal
function closeModal() {
    document.getElementById('seasonModal').classList.remove('active');
}

// Zapisz sezon
document.getElementById('seasonForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const id = document.getElementById('seasonId').value;
    const data = {
        number: parseInt(document.getElementById('seasonNumber').value),
        description: document.getElementById('seasonDescription').value,
        is_current: document.getElementById('seasonIsCurrent').checked
    };

    try {
        const url = id ? `/api/seasons/${id}` : '/api/seasons';
        const method = id ? 'PUT' : 'POST';

        const response = await fetch(url, {
            method: method,
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(data)
        });

        if (response.ok) {
            closeModal();
            loadSeasons();
        } else {
            alert('Błąd zapisu sezonu');
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
});

// Ustaw aktualny sezon
async function setCurrentSeason(id) {
    if (!confirm('Czy na pewno chcesz ustawić ten sezon jako aktualny?')) return;

    try {
        const response = await fetch(`/api/seasons/${id}/set-current`, {
            method: 'POST'
        });

        if (response.ok) {
            loadSeasons();
        } else {
            alert('Błąd ustawiania sezonu');
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

// Usuń sezon
async function deleteSeason(id) {
    if (!confirm('Czy na pewno chcesz usunąć ten sezon? Ta operacja jest nieodwracalna.')) return;

    try {
        const response = await fetch(`/api/seasons/${id}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            loadSeasons();
        } else {
            alert('Błąd usuwania sezonu');
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

// Inicjalizacja
loadSeasons();