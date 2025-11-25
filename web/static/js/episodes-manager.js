// Global state
let episodes = [];
let seasons = [];
let allStaff = [];
let allGuests = [];
let staffTypes = [];
let guestTypes = [];
let sources = [];
let mediaSceneId = null;
let reportazeSceneId = null;
let currentEpisodeId = null;
let assignedStaff = [];
let assignedGuests = [];
let assignedMedia = [];
let availableMediaFiles = [];
let mediaGroups = [];
let currentMediaGroup = null;

// ===== INITIALIZATION =====
document.addEventListener('DOMContentLoaded', () => {
    loadSeasons();
    loadEpisodes();
    loadStaffTypes();
    loadGuestTypes();
    loadMediaScenes();
    setupEventListeners();
});

function setupEventListeners() {
    // Episode form
    document.getElementById('episodeForm').addEventListener('submit', (e) => {
        e.preventDefault();
    });

    // Add staff form
    document.getElementById('addStaffForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        await createStaff();
    });

    // Add staff type form
    document.getElementById('addStaffTypeForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        await createStaffType();
    });

    // Edit staff types form
    document.getElementById('editStaffTypesForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        await updateStaffTypes();
    });

    // Add guest form
    document.getElementById('addGuestForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        await createGuest();
    });

    // Add guest type form
    document.getElementById('addGuestTypeForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        await createGuestType();
    });

    // Assign media form
    document.getElementById('assignMediaForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        await assignMedia();
    });

    // Edit guest form
    document.getElementById('editGuestForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        await updateGuestAssignment();
    });

    // Add media group form
    document.getElementById('addMediaGroupForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        await createMediaGroup();
    });
}

// ===== TAB SWITCHING =====
function switchTab(tabName, sourceElement) {
    // Update tab buttons
    document.querySelectorAll('.modal-tab').forEach(tab => {
        tab.classList.remove('active');
    });
    
    // Jeśli podano element źródłowy, ustaw go jako aktywny
    if (sourceElement) {
        sourceElement.classList.add('active');
    } else {
        // Jeśli nie, znajdź zakładkę odpowiadającą tabName
        const tabButtons = document.querySelectorAll('.modal-tab');
        const tabIndex = ['data', 'staff', 'guests', 'media'].indexOf(tabName);
        if (tabIndex >= 0 && tabButtons[tabIndex]) {
            tabButtons[tabIndex].classList.add('active');
        }
    }

    // Update tab content
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });
    document.getElementById('tab' + tabName.charAt(0).toUpperCase() + tabName.slice(1)).classList.add('active');

    // Load data when switching to specific tabs
    if (tabName === 'staff') {
        loadAllStaff();
        if (currentEpisodeId) loadAssignedStaff();
    } else if (tabName === 'guests') {
        loadAllGuests();
        if (currentEpisodeId) loadAssignedGuests();
    } else if (tabName === 'media') {
        if (currentEpisodeId) {
            loadMediaFiles();
            loadAssignedMedia();
            loadMediaGroups();
            // Reload staff dla opcji autora
            loadAssignedStaff().then(() => updateMediaStaffSelect());
        }
    }
}

// ===== MEDIA SUB-TAB SWITCHING =====
function switchMediaSubTab(subTabName, sourceElement) {
    // Ukryj wszystkie pod-zakładki
    document.querySelectorAll('.sub-tab-content').forEach(content => {
        content.classList.remove('active');
    });
    
    // Usuń active z przycisków pod-zakładek
    document.querySelectorAll('.sub-tab').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Pokaż wybraną pod-zakładkę
    const contentId = 'mediaSubTab' + subTabName.charAt(0).toUpperCase() + subTabName.slice(1);
    document.getElementById(contentId).classList.add('active');
    
    // Zaznacz przycisk
    if (sourceElement) {
        sourceElement.classList.add('active');
    }
}

// ===== SEASONS =====
async function loadSeasons() {
    try {
        const response = await fetch('/api/seasons');
        seasons = await response.json();
        updateSeasonSelects();
    } catch (error) {
        console.error('Błąd ładowania sezonów:', error);
    }
}

function updateSeasonSelects() {
    // Filter
    const filterSelect = document.getElementById('seasonFilter');
    filterSelect.innerHTML = '<option value="">Wszystkie</option>' +
        seasons.map(s => `<option value="${s.id}">Sezon ${s.number}</option>`).join('');

    // Modal
    const modalSelect = document.getElementById('episodeSeason');
    modalSelect.innerHTML = '<option value="">Wybierz sezon...</option>' +
        seasons.map(s => `<option value="${s.id}">Sezon ${s.number}${s.is_current ? ' (aktualny)' : ''}</option>`).join('');
}

// ===== EPISODES =====
async function loadEpisodes() {
    try {
        const seasonId = document.getElementById('seasonFilter').value;
        const url = seasonId ? `/api/episodes?season_id=${seasonId}` : '/api/episodes';
        const response = await fetch(url);
        episodes = await response.json();
        renderEpisodes();
    } catch (error) {
        console.error('Błąd ładowania odcinków:', error);
    }
}

function renderEpisodes() {
    const tbody = document.getElementById('episodesTableBody');
    
    if (episodes.length === 0) {
        tbody.innerHTML = `
            <tr>
                <td colspan="6">
                    <div class="empty-state">
                        <div class="empty-state-icon">📺</div>
                        <div>Brak odcinków. Utwórz pierwszy odcinek.</div>
                    </div>
                </td>
            </tr>
        `;
        return;
    }

    tbody.innerHTML = episodes.map(episode => {
        // Formatuj datę jako YYYY-MM-DD
        let date = '-';
        if (episode.episode_date) {
            const d = new Date(episode.episode_date);
            const yyyy = d.getFullYear();
            const mm = String(d.getMonth() + 1).padStart(2, '0');
            const dd = String(d.getDate()).padStart(2, '0');
            date = `${yyyy}-${mm}-${dd}`;
        }
        const season = episode.season ? episode.season.number : '-';
        
        return `
            <tr>
                <td><strong>${episode.episode_number}</strong></td>
                <td>S${season}E${episode.season_episode}</td>
                <td>${episode.title}</td>
                <td>${date}</td>
                <td>
                    ${episode.is_current ? '<span class="badge badge-success">Aktualny</span>' : '<span class="badge badge-secondary">-</span>'}
                </td>
                <td>
                    <div class="table-actions">
                        ${!episode.is_current ? `<button class="btn btn-success btn-small" onclick="setCurrentEpisode(${episode.id})">Aktywuj</button>` : ''}
                        <button class="btn btn-primary btn-small" onclick="openEditModal(${episode.id})">Edytuj</button>
                        <button class="btn btn-danger btn-small" onclick="deleteEpisode(${episode.id})">Usuń</button>
                    </div>
                </td>
            </tr>
        `;
    }).join('');
}

async function openCreateModal() {
    document.getElementById('modalTitle').textContent = 'Nowy Odcinek';
    document.getElementById('episodeForm').reset();
    document.getElementById('episodeId').value = '';
    currentEpisodeId = null;
    assignedStaff = [];
    assignedGuests = [];
    assignedMedia = [];
    
    // Pobierz następne numery odcinków
    try {
        const response = await fetch('/api/episodes/next-numbers');
        const data = await response.json();
        
        // Ustaw aktualny sezon
        if (data.current_season_id) {
            document.getElementById('episodeSeason').value = data.current_season_id;
        }
        
        // Ustaw numery
        document.getElementById('episodeNumber').value = data.next_episode_number;
        document.getElementById('seasonEpisode').value = data.next_season_episode;
    } catch (error) {
        console.error('Błąd pobierania następnych numerów:', error);
        document.getElementById('episodeNumber').value = 1;
        document.getElementById('seasonEpisode').value = 1;
    }
    
    // Ustaw dzisiejszą datę w formacie YYYY-MM-DD
    const today = new Date();
    const yyyy = today.getFullYear();
    const mm = String(today.getMonth() + 1).padStart(2, '0');
    const dd = String(today.getDate()).padStart(2, '0');
    document.getElementById('episodeDate').value = `${yyyy}-${mm}-${dd}`;
    
    // Switch to first tab
    switchTab('data');
    
    document.getElementById('episodeModal').classList.add('active');
}

function openEditModal(id) {
    const episode = episodes.find(e => e.id === id);
    if (!episode) return;

    currentEpisodeId = id;
    document.getElementById('modalTitle').textContent = 'Edycja Odcinka';
    document.getElementById('episodeId').value = episode.id;
    document.getElementById('episodeSeason').value = episode.season_id;
    document.getElementById('episodeNumber').value = episode.episode_number;
    document.getElementById('seasonEpisode').value = episode.season_episode;
    document.getElementById('episodeTitle').value = episode.title;
    
    if (episode.episode_date) {
        const date = new Date(episode.episode_date);
        document.getElementById('episodeDate').value = date.toISOString().split('T')[0];
    }
    
    document.getElementById('episodeIsCurrent').checked = episode.is_current;
    
    // Switch to first tab
    switchTab('data');
    
    document.getElementById('episodeModal').classList.add('active');
}

function closeModal() {
    document.getElementById('episodeModal').classList.remove('active');
    currentEpisodeId = null;
}

async function saveEpisode() {
    const id = document.getElementById('episodeId').value;
    const dateValue = document.getElementById('episodeDate').value;
    
    // Przygotuj datę - jeśli nie podano, użyj dzisiejszej w formacie YYYY-MM-DD
    let episodeDate = dateValue;
    if (!episodeDate) {
        const today = new Date();
        const yyyy = today.getFullYear();
        const mm = String(today.getMonth() + 1).padStart(2, '0');
        const dd = String(today.getDate()).padStart(2, '0');
        episodeDate = `${yyyy}-${mm}-${dd}`;
    }
    
    // Walidacja formatu daty YYYY-MM-DD
    const datePattern = /^\d{4}-\d{2}-\d{2}$/;
    if (!datePattern.test(episodeDate)) {
        alert('Data musi być w formacie YYYY-MM-DD (np. 2024-12-31)');
        return;
    }
    
    const data = {
        season_id: parseInt(document.getElementById('episodeSeason').value),
        episode_number: parseInt(document.getElementById('episodeNumber').value),
        season_episode: parseInt(document.getElementById('seasonEpisode').value),
        title: document.getElementById('episodeTitle').value,
        episode_date: episodeDate + 'T00:00:00Z', // Dodaj czas dla kompatybilności z backendem
        is_current: document.getElementById('episodeIsCurrent').checked
    };

    try {
        const url = id ? `/api/episodes/${id}` : '/api/episodes';
        const method = id ? 'PUT' : 'POST';
        
        const response = await fetch(url, {
            method: method,
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(data)
        });

        if (response.ok) {
            closeModal();
            loadEpisodes();
        } else {
            const error = await response.text();
            alert('Błąd zapisu odcinka: ' + error);
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

async function setCurrentEpisode(id) {
    if (!confirm('Czy na pewno chcesz ustawić ten odcinek jako aktualny?')) return;

    try {
        const response = await fetch(`/api/episodes/${id}/set-current`, {
            method: 'POST'
        });

        if (response.ok) {
            loadEpisodes();
        } else {
            alert('Błąd ustawiania odcinka');
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

async function deleteEpisode(id) {
    if (!confirm('Czy na pewno chcesz usunąć ten odcinek? Ta operacja jest nieodwracalna.')) return;

    try {
        const response = await fetch(`/api/episodes/${id}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            loadEpisodes();
        } else {
            const error = await response.text();
            alert('Błąd usuwania odcinka: ' + error);
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

// ===== STAFF =====
async function loadStaffTypes() {
    try {
        const response = await fetch('/api/staff-types');
        staffTypes = await response.json();
        updateStaffTypeSelect();
    } catch (error) {
        console.error('Błąd ładowania typów staff:', error);
    }
}

function updateStaffTypeSelect() {
    const select = document.getElementById('staffType');
    if (!select) return; // Element nie istnieje w obecnym kontekście
    select.innerHTML = '<option value="">Wybierz typ...</option>' +
        staffTypes.map(t => `<option value="${t.id}">${t.name}</option>`).join('');
}

async function loadAllStaff() {
    try {
        const response = await fetch('/api/staff');
        allStaff = await response.json();
        renderAvailableStaff();
    } catch (error) {
        console.error('Błąd ładowania staff:', error);
    }
}

function renderAvailableStaff() {
    const container = document.getElementById('availableStaffList');
    
    if (allStaff.length === 0) {
        container.innerHTML = '<div style="text-align: center; padding: 20px; color: #666;">Brak dostępnych</div>';
        return;
    }

    // Filter out already assigned
    const assignedIds = assignedStaff.map(s => s.staff_id);
    const available = allStaff.filter(s => !assignedIds.includes(s.id));

    if (available.length === 0) {
        container.innerHTML = '<div style="text-align: center; padding: 20px; color: #666;">Wszyscy przypisani</div>';
        return;
    }

    container.innerHTML = available.map(staff => `
        <div class="list-item">
            <div class="list-item-info">
                <div>${staff.first_name} ${staff.last_name}</div>
            </div>
            <div class="list-item-actions">
                <button class="btn btn-success btn-icon" onclick="assignStaffToEpisode(${staff.id})">+</button>
            </div>
        </div>
    `).join('');
}

async function loadAssignedStaff() {
    if (!currentEpisodeId) return;
    
    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/staff`);
        assignedStaff = await response.json();
        renderAssignedStaff();
        renderAvailableStaff();
    } catch (error) {
        console.error('Błąd ładowania przypisanego staff:', error);
    }
}

function renderAssignedStaff() {
    const container = document.getElementById('assignedStaffList');
    
    if (assignedStaff.length === 0) {
        container.innerHTML = '<div style="text-align: center; padding: 20px; color: #666;">Brak przypisanych</div>';
        return;
    }

    container.innerHTML = assignedStaff.map(assignment => {
        const staff = assignment.staff;
        const types = assignment.staff_types ? 
            assignment.staff_types.map(st => st.staff_type.name).join(', ') : 
            'Brak typu';
        
        return `
            <div class="list-item">
                <div class="list-item-info">
                    <div>${staff.first_name} ${staff.last_name}</div>
                    <div class="list-item-type">${types}</div>
                </div>
                <div class="list-item-actions">
                    <button class="btn btn-primary btn-icon" onclick="openEditStaffTypesModal(${assignment.id})">✎</button>
                    <button class="btn btn-danger btn-icon" onclick="removeStaffFromEpisode(${assignment.id})">×</button>
                </div>
            </div>
        `;
    }).join('');
}

async function assignStaffToEpisode(staffId) {
    if (!currentEpisodeId) {
        alert('Najpierw zapisz odcinek');
        return;
    }

    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/staff`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ staff_id: staffId })
        });

        if (response.ok) {
            await loadAssignedStaff();
        } else {
            const error = await response.text();
            alert('Błąd przypisywania: ' + error);
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

async function removeStaffFromEpisode(assignmentId) {
    if (!confirm('Czy na pewno chcesz usunąć to przypisanie?')) return;

    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/staff/${assignmentId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            await loadAssignedStaff();
        } else {
            alert('Błąd usuwania przypisania');
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

function openAddStaffModal() {
    document.getElementById('addStaffForm').reset();
    document.getElementById('addStaffModal').classList.add('active');
}

function closeAddStaffModal() {
    document.getElementById('addStaffModal').classList.remove('active');
}

async function createStaff() {
    const data = {
        first_name: document.getElementById('staffFirstName').value,
        last_name: document.getElementById('staffLastName').value
    };

    try {
        const response = await fetch('/api/staff', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(data)
        });

        if (response.ok) {
            closeAddStaffModal();
            await loadAllStaff();
        } else {
            const error = await response.text();
            alert('Błąd dodawania: ' + error);
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

function openAddStaffTypeModal() {
    document.getElementById('addStaffTypeForm').reset();
    document.getElementById('addStaffTypeModal').classList.add('active');
}

function openAddStaffTypeModalFromEdit() {
    // Zapamiętaj że otwieramy z edycji
    window.staffTypeFromEdit = true;
    openAddStaffTypeModal();
}

function closeAddStaffTypeModal() {
    document.getElementById('addStaffTypeModal').classList.remove('active');
    // Jeśli był otwarty z edycji typów, odśwież listę i wróć do modala edycji
    if (window.staffTypeFromEdit) {
        window.staffTypeFromEdit = false;
        loadStaffTypes().then(() => {
            // Odśwież select w modalu edycji
            const assignmentId = document.getElementById('editStaffAssignmentId').value;
            if (assignmentId) {
                openEditStaffTypesModal(parseInt(assignmentId));
            }
        });
    }
}

async function createStaffType() {
    const data = {
        name: document.getElementById('staffTypeName').value
    };

    try {
        const response = await fetch('/api/staff-types', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(data)
        });

        if (response.ok) {
            closeAddStaffTypeModal();
            await loadStaffTypes();
        } else {
            const error = await response.text();
            alert('Błąd dodawania typu: ' + error);
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

function openEditStaffTypesModal(assignmentId) {
    const assignment = assignedStaff.find(a => a.id === assignmentId);
    if (!assignment) return;

    document.getElementById('editStaffAssignmentId').value = assignmentId;
    document.getElementById('editStaffName').textContent = 
        `${assignment.staff.first_name} ${assignment.staff.last_name}`;
    
    // Wypełnij select typami
    const select = document.getElementById('editStaffTypesSelect');
    select.innerHTML = staffTypes.map(type => 
        `<option value="${type.id}">${type.name}</option>`
    ).join('');
    
    // Zaznacz przypisane typy
    const assignedTypeIds = assignment.staff_types ? 
        assignment.staff_types.map(st => st.staff_type_id) : [];
    
    Array.from(select.options).forEach(option => {
        option.selected = assignedTypeIds.includes(parseInt(option.value));
    });
    
    document.getElementById('editStaffTypesModal').classList.add('active');
}

function closeEditStaffTypesModal() {
    document.getElementById('editStaffTypesModal').classList.remove('active');
}

async function updateStaffTypes() {
    const assignmentId = document.getElementById('editStaffAssignmentId').value;
    const select = document.getElementById('editStaffTypesSelect');
    const selectedTypes = Array.from(select.selectedOptions).map(opt => parseInt(opt.value));

    const data = {
        staff_type_ids: selectedTypes
    };

    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/staff/${assignmentId}/types`, {
            method: 'PUT',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(data)
        });

        if (response.ok) {
            closeEditStaffTypesModal();
            await loadAssignedStaff();
        } else {
            const error = await response.text();
            alert('Błąd aktualizacji: ' + error);
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

// ===== GUESTS =====
async function loadGuestTypes() {
    try {
        const response = await fetch('/api/guest-types');
        guestTypes = await response.json();
        updateGuestTypeSelect();
    } catch (error) {
        console.error('Błąd ładowania typów gości:', error);
    }
}

function updateGuestTypeSelect() {
    const select = document.getElementById('guestType');
    select.innerHTML = '<option value="">Wybierz typ...</option>' +
        guestTypes.map(t => `<option value="${t.id}">${t.name}</option>`).join('');
}

async function loadAllGuests() {
    try {
        const response = await fetch('/api/guests');
        allGuests = await response.json();
        renderAvailableGuests();
    } catch (error) {
        console.error('Błąd ładowania gości:', error);
    }
}

function renderAvailableGuests() {
    const container = document.getElementById('availableGuestsList');
    
    if (allGuests.length === 0) {
        container.innerHTML = '<div style="text-align: center; padding: 20px; color: #666;">Brak dostępnych</div>';
        return;
    }

    // Filter out already assigned
    const assignedIds = assignedGuests.map(g => g.guest_id);
    const available = allGuests.filter(g => !assignedIds.includes(g.id));

    if (available.length === 0) {
        container.innerHTML = '<div style="text-align: center; padding: 20px; color: #666;">Wszyscy przypisani</div>';
        return;
    }

    container.innerHTML = available.map(guest => `
        <div class="list-item">
            <div class="list-item-info">
                <div>${guest.first_name} ${guest.last_name}</div>
                <div class="list-item-type">${guest.guest_type ? guest.guest_type.name : ''}</div>
            </div>
            <div class="list-item-actions">
                <button class="btn btn-success btn-icon" onclick="assignGuestToEpisode(${guest.id})">+</button>
            </div>
        </div>
    `).join('');
}

async function loadAssignedGuests() {
    if (!currentEpisodeId) return;
    
    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/guests`);
        assignedGuests = await response.json();
        renderAssignedGuests();
        renderAvailableGuests();
    } catch (error) {
        console.error('Błąd ładowania przypisanych gości:', error);
    }
}

function renderAssignedGuests() {
    const container = document.getElementById('assignedGuestsList');
    
    if (assignedGuests.length === 0) {
        container.innerHTML = '<div style="text-align: center; padding: 20px; color: #666;">Brak przypisanych</div>';
        return;
    }

    container.innerHTML = assignedGuests.map(assignment => {
        const guest = assignment.guest;
        return `
            <div class="list-item">
                <div class="list-item-info">
                    <div>${guest.first_name} ${guest.last_name}</div>
                    <div class="list-item-type">${guest.guest_type ? guest.guest_type.name : ''}</div>
                    ${assignment.topic ? `<div class="list-item-type">Temat: ${assignment.topic}</div>` : ''}
                    ${assignment.segment_order ? `<div class="list-item-type">Kolejność: ${assignment.segment_order}</div>` : ''}
                </div>
                <div class="list-item-actions">
                    <button class="btn btn-primary btn-icon" onclick="openEditGuestModal(${assignment.id})">✎</button>
                    <button class="btn btn-danger btn-icon" onclick="removeGuestFromEpisode(${assignment.id})">×</button>
                </div>
            </div>
        `;
    }).join('');
}

async function assignGuestToEpisode(guestId) {
    if (!currentEpisodeId) {
        alert('Najpierw zapisz odcinek');
        return;
    }

    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/guests`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ 
                guest_id: guestId,
                topic: '',
                segment_order: assignedGuests.length + 1
            })
        });

        if (response.ok) {
            await loadAssignedGuests();
        } else {
            const error = await response.text();
            alert('Błąd przypisywania: ' + error);
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

async function removeGuestFromEpisode(assignmentId) {
    if (!confirm('Czy na pewno chcesz usunąć to przypisanie?')) return;

    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/guests/${assignmentId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            await loadAssignedGuests();
        } else {
            alert('Błąd usuwania przypisania');
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

function openAddGuestModal() {
    document.getElementById('addGuestForm').reset();
    document.getElementById('addGuestModal').classList.add('active');
}

function closeAddGuestModal() {
    document.getElementById('addGuestModal').classList.remove('active');
}

async function createGuest() {
    const typeId = document.getElementById('guestType').value;
    const data = {
        first_name: document.getElementById('guestFirstName').value,
        last_name: document.getElementById('guestLastName').value
    };
    
    if (typeId) {
        data.guest_type_id = parseInt(typeId);
    }

    try {
        const response = await fetch('/api/guests', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(data)
        });

        if (response.ok) {
            closeAddGuestModal();
            await loadAllGuests();
        } else {
            const error = await response.text();
            alert('Błąd dodawania: ' + error);
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

function openAddGuestTypeModal() {
    document.getElementById('addGuestTypeForm').reset();
    document.getElementById('addGuestTypeModal').classList.add('active');
}

function closeAddGuestTypeModal() {
    document.getElementById('addGuestTypeModal').classList.remove('active');
}

async function createGuestType() {
    const data = {
        name: document.getElementById('guestTypeName').value
    };

    try {
        const response = await fetch('/api/guest-types', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(data)
        });

        if (response.ok) {
            closeAddGuestTypeModal();
            await loadGuestTypes();
        } else {
            const error = await response.text();
            alert('Błąd dodawania typu: ' + error);
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

function openEditGuestModal(assignmentId) {
    const assignment = assignedGuests.find(a => a.id === assignmentId);
    if (!assignment) return;

    document.getElementById('editGuestAssignmentId').value = assignmentId;
    document.getElementById('editGuestName').textContent = 
        `${assignment.guest.first_name} ${assignment.guest.last_name}`;
    document.getElementById('editGuestTopic').value = assignment.topic || '';
    document.getElementById('editGuestOrder').value = assignment.segment_order || 1;
    
    document.getElementById('editGuestModal').classList.add('active');
}

function closeEditGuestModal() {
    document.getElementById('editGuestModal').classList.remove('active');
}

async function updateGuestAssignment() {
    const assignmentId = document.getElementById('editGuestAssignmentId').value;
    const data = {
        topic: document.getElementById('editGuestTopic').value,
        segment_order: parseInt(document.getElementById('editGuestOrder').value)
    };

    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/guests/${assignmentId}`, {
            method: 'PUT',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(data)
        });

        if (response.ok) {
            closeEditGuestModal();
            await loadAssignedGuests();
        } else {
            const error = await response.text();
            alert('Błąd aktualizacji: ' + error);
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

// ===== MEDIA =====
async function loadMediaScenes() {
    try {
        const response = await fetch('/api/scenes/media');
        const scenes = await response.json();
        
        // Store scenes
        window.mediaScenes = scenes;
        
        // Zapisz ID scen
        const mediaScene = scenes.find(s => s.name === 'MEDIA');
        const reportazeScene = scenes.find(s => s.name === 'REPORTAZE');
        
        if (mediaScene) mediaSceneId = mediaScene.id;
        if (reportazeScene) reportazeSceneId = reportazeScene.id;
        
        updateMediaStaffSelect();
    } catch (error) {
        console.error('Błąd ładowania scen:', error);
    }
}

function updateMediaStaffSelect() {
    const select = document.getElementById('mediaStaff');
    // Wypełnij przypisanymi członkami ekipy
    select.innerHTML = '<option value="">Brak</option>' +
        assignedStaff.map(assignment => 
            `<option value="${assignment.id}">${assignment.staff.first_name} ${assignment.staff.last_name}</option>`
        ).join('');
}

function updateMediaGroupSelect() {
    const container = document.getElementById('mediaGroupCheckboxes');
    if (!container) return; // Element nie istnieje jeśli formularz nie jest otwarty
    
    if (mediaGroups.length === 0) {
        container.innerHTML = '<div style="text-align: center; color: #666; font-size: 11px;">Brak dostępnych grup</div>';
        return;
    }
    
    // Grupuj według nazwy - traktuj grupy o tej samej nazwie jako jedną
    const groupsByName = {};
    mediaGroups.forEach(group => {
        if (!groupsByName[group.name]) {
            groupsByName[group.name] = {
                name: group.name,
                ids: [],
                scenes: []
            };
        }
        groupsByName[group.name].ids.push(group.id);
        if (group.scene) {
            groupsByName[group.name].scenes.push(group.scene.name);
        }
    });
    
    // Renderuj unikalne grupy
    const uniqueGroups = Object.values(groupsByName);
    
    container.innerHTML = uniqueGroups.map(group => {
        // Deduplikuj sceny
        const uniqueScenes = [...new Set(group.scenes)];
        const sceneLabel = uniqueScenes.length > 0 ? ` (${uniqueScenes.join(', ')})` : '';
        const groupIdsStr = group.ids.join(','); // Przechowuj wszystkie ID jako string
        
        return `<label style="display: flex; align-items: center; gap: 5px; margin-bottom: 5px; cursor: pointer; padding: 4px; border-radius: 3px; transition: background 0.2s;" onmouseover="this.style.background='rgba(255,255,255,0.05)'" onmouseout="this.style.background='transparent'">
            <input type="checkbox" class="media-group-checkbox" value="${groupIdsStr}" style="cursor: pointer;">
            <span style="font-size: 11px;">${group.name}<span style="color: #888; font-size: 9px;">${sceneLabel}</span></span>
        </label>`;
    }).join('');
}

async function loadMediaFiles() {
    if (!currentEpisodeId) return;
    
    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/media/files`);
        availableMediaFiles = await response.json();
        renderMediaFiles();
    } catch (error) {
        console.error('Błąd ładowania plików:', error);
    }
}

function renderMediaFiles() {
    const container = document.getElementById('mediaFilesGrid');
    
    if (availableMediaFiles.length === 0) {
        container.innerHTML = '<div style="grid-column: 1/-1; text-align: center; padding: 20px; color: #666;">Brak plików w folderze sezonu</div>';
        return;
    }

    // Pobierz ścieżki już przypisanych plików
    const assignedFilePaths = assignedMedia.map(m => m.file_path).filter(Boolean);

    container.innerHTML = availableMediaFiles.map(file => {
        const isAssigned = assignedFilePaths.includes(file.path);
        const assignedClass = isAssigned ? 'assigned' : '';
        const assignedBadge = isAssigned ? '<span class="badge badge-success" style="font-size: 8px; margin-left: 5px;">PRZYPISANY</span>' : '';
        
        return `
            <div class="media-file-card ${assignedClass}" onclick="${isAssigned ? '' : `selectMediaFile('${file.path}', '${file.name}', ${file.duration})`}" style="${isAssigned ? 'opacity: 0.5; cursor: not-allowed;' : ''}">
                <div class="media-file-name">${file.name}${assignedBadge}</div>
                <div class="media-file-info">
                    Typ: ${file.type}<br>
                    ${file.duration ? `Czas: ${formatDuration(file.duration)}` : ''}
                </div>
            </div>
        `;
    }).join('');
}

function selectMediaFile(path, name, duration) {
    document.getElementById('mediaFilePath').value = path;
    document.getElementById('mediaFileDuration').value = duration || 0;
    document.getElementById('mediaFileName').textContent = name;
    document.getElementById('mediaTitle').value = name.replace(/\.[^/.]+$/, ''); // Remove extension
    updateMediaGroupSelect(); // Wypełnij select grup mediów
    openAssignMediaModal();
}

function openAssignMediaModal() {
    document.getElementById('sceneMedia').checked = false;
    document.getElementById('sceneReportaze').checked = false;
    document.getElementById('sceneError').style.display = 'none';
    document.getElementById('assignMediaModal').classList.add('active');
}

function closeAssignMediaModal() {
    document.getElementById('assignMediaModal').classList.remove('active');
    document.getElementById('assignMediaForm').reset();
}

async function assignMedia() {
    if (!currentEpisodeId) {
        alert('Najpierw zapisz odcinek');
        return;
    }

    // Walidacja scen
    const mediaChecked = document.getElementById('sceneMedia').checked;
    const reportazeChecked = document.getElementById('sceneReportaze').checked;
    
    if (!mediaChecked && !reportazeChecked) {
        document.getElementById('sceneError').style.display = 'block';
        return;
    }
    
    const filePath = document.getElementById('mediaFilePath').value;
    const staffId = document.getElementById('mediaStaff').value;
    
    // Zbierz zaznaczone grupy - wartość może być pojedynczym ID lub listą ID oddzieloną przecinkami
    const selectedGroupIds = Array.from(document.querySelectorAll('.media-group-checkbox:checked'))
        .flatMap(cb => {
            // Wartość może być "123" lub "123,456"
            return cb.value.split(',').map(id => parseInt(id.trim()));
        })
        .filter(id => !isNaN(id)); // Usuń nieprawidłowe wartości
    
    const modal = document.getElementById('assignMediaModal');
    const isEditMode = modal.dataset.editMode === 'true';
    const originalFilePath = modal.dataset.originalFilePath;
    
    const baseData = {
        title: document.getElementById('mediaTitle').value,
        description: document.getElementById('mediaDescription').value,
        file_path: filePath,
        duration: parseInt(document.getElementById('mediaFileDuration').value),
        episode_staff_id: staffId ? parseInt(staffId) : null,
        order: 0
    };

    try {
        if (isEditMode) {
            // Edycja - usuń stare wpisy, dodaj nowe
            const oldMediaItems = assignedMedia.filter(m => m.file_path === originalFilePath);
            
            for (const item of oldMediaItems) {
                await fetch(`/api/episodes/${currentEpisodeId}/media/${item.id}`, {
                    method: 'DELETE'
                });
            }
        }
        
        const selectedScenes = [];
        if (mediaChecked) selectedScenes.push({name: 'MEDIA', id: mediaSceneId});
        if (reportazeChecked) selectedScenes.push({name: 'REPORTAZE', id: reportazeSceneId});
        
        const createdMediaIds = []; // Zbieramy ID utworzonych mediów
        
        // Utwórz wpis dla każdej wybranej sceny
        for (const scene of selectedScenes) {
            const data = { ...baseData, scene_id: scene.id };
            
            const response = await fetch(`/api/episodes/${currentEpisodeId}/media`, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(data)
            });

            if (response.ok) {
                const newMedia = await response.json();
                createdMediaIds.push(newMedia.id);
            } else if (response.status === 409) {
                alert(`Ten plik jest już przypisany do sceny ${scene.name} w tym odcinku`);
                return;
            } else {
                const error = await response.text();
                alert('Błąd przypisywania media: ' + error);
                return;
            }
        }
        
        // Dodaj wszystkie utworzone media do wybranych grup (tylko raz)
        if (selectedGroupIds.length > 0 && createdMediaIds.length > 0) {
            for (const groupId of selectedGroupIds) {
                for (const mediaId of createdMediaIds) {
                    try {
                        const response = await fetch(`/api/media-groups/${groupId}/items`, {
                            method: 'POST',
                            headers: {'Content-Type': 'application/json'},
                            body: JSON.stringify({
                                episode_media_id: mediaId,
                                order: 0
                            })
                        });
                        
                        // Ignoruj błąd 409 (Conflict) - media już jest w grupie
                        if (!response.ok && response.status !== 409) {
                            console.error('Błąd dodawania do grupy:', await response.text());
                        }
                    } catch (error) {
                        console.error('Błąd dodawania do grupy:', error);
                    }
                }
            }
        }
        
        // Wyczyść tryb edycji
        delete modal.dataset.editMode;
        delete modal.dataset.originalFilePath;
        
        closeAssignMediaModal();
        await loadAssignedMedia();
        await loadMediaFiles(); // Odśwież listę plików
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

async function loadAssignedMedia() {
    if (!currentEpisodeId) return;
    
    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/media`);
        assignedMedia = await response.json();
        renderAssignedMedia();
    } catch (error) {
        console.error('Błąd ładowania przypisanych mediów:', error);
    }
}

function renderAssignedMedia() {
    const containerMedia = document.getElementById('assignedMediaListMedia');
    const containerReportaze = document.getElementById('assignedMediaListReportaze');
    
    if (assignedMedia.length === 0) {
        containerMedia.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak mediów MEDIA</div>';
        containerReportaze.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak mediów REPORTAŻE</div>';
        return;
    }

    // Grupuj media po file_path i scenie
    const mediaByFileAndScene = {};
    assignedMedia.forEach(m => {
        const sceneName = m.scene ? m.scene.name : 'UNKNOWN';
        const key = `${m.file_path}_${sceneName}`;
        
        if (!mediaByFileAndScene[key]) {
            mediaByFileAndScene[key] = {
                title: m.title,
                file_path: m.file_path,
                description: m.description,
                duration: m.duration,
                author: m.episode_staff,
                scene: m.scene,
                mediaItems: []
            };
        }
        mediaByFileAndScene[key].mediaItems.push(m);
    });
    
    const groupedMedia = Object.values(mediaByFileAndScene);
    
    // Rozdziel na MEDIA i REPORTAZE
    const mediaItems_MEDIA = groupedMedia.filter(m => m.scene?.name === 'MEDIA');
    const mediaItems_REPORTAZE = groupedMedia.filter(m => m.scene?.name === 'REPORTAZE');
    
    // Funkcja renderująca pojedynczy element
    const renderMediaItem = (media) => {
        const authorName = media.author && media.author.staff ? 
            `${media.author.staff.first_name} ${media.author.staff.last_name}` : 
            'Brak';
        
        // Sprawdź czy któryś z wpisów jest current
        const hasCurrent = media.mediaItems.some(m => m.is_current);
        const currentBadge = hasCurrent ? 
            '<span class="badge badge-success">WCZYTANY</span>' : '';
        
        return `
            <div class="assigned-media-item" onclick="editMediaAssignment('${media.file_path}')">
                <div class="media-item-details">
                    <div class="media-item-title">${media.title} ${currentBadge}</div>
                    <div class="media-item-meta">
                        Autor: ${authorName}<br>
                        ${media.description ? `Opis: ${media.description}<br>` : ''}
                        Plik: ${media.file_path || 'Brak'}<br>
                        ${media.duration ? `Czas: ${formatDuration(media.duration)}<br>` : ''}
                    </div>
                </div>
                <div class="list-item-actions" style="pointer-events: auto;">
                    <button class="btn btn-primary btn-icon" onclick="event.stopPropagation(); editMediaAssignment('${media.file_path}')" title="Edytuj">✎</button>
                    <button class="btn btn-danger btn-icon" onclick="event.stopPropagation(); removeMediaFile('${media.file_path}')" title="Usuń">×</button>
                </div>
            </div>
        `;
    };
    
    // Renderuj MEDIA
    if (mediaItems_MEDIA.length === 0) {
        containerMedia.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak mediów MEDIA</div>';
    } else {
        containerMedia.innerHTML = mediaItems_MEDIA.map(renderMediaItem).join('');
    }
    
    // Renderuj REPORTAZE
    if (mediaItems_REPORTAZE.length === 0) {
        containerReportaze.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak mediów REPORTAŻE</div>';
    } else {
        containerReportaze.innerHTML = mediaItems_REPORTAZE.map(renderMediaItem).join('');
    }
}
//            </div>
//       `;
//    }).join('');
//}

function editMediaAssignment(filePath) {
    // Znajdź wszystkie media z tym plikiem
    const mediaItems = assignedMedia.filter(m => m.file_path === filePath);
    if (mediaItems.length === 0) return;

    const first = mediaItems[0];
    
    // Wypełnij formularz
    document.getElementById('mediaFilePath').value = filePath;
    document.getElementById('mediaFileDuration').value = first.duration || 0;
    document.getElementById('mediaFileName').textContent = filePath.split('/').pop();
    document.getElementById('mediaTitle').value = first.title;
    document.getElementById('mediaDescription').value = first.description || '';
    document.getElementById('mediaStaff').value = first.episode_staff_id || '';
    
    // Zaznacz sceny
    document.getElementById('sceneMedia').checked = mediaItems.some(m => m.scene?.name === 'MEDIA');
    document.getElementById('sceneReportaze').checked = mediaItems.some(m => m.scene?.name === 'REPORTAZE');
    document.getElementById('sceneError').style.display = 'none';
    
    // Zapisz oryginalne file_path jako identyfikator edycji
    document.getElementById('assignMediaModal').dataset.editMode = 'true';
    document.getElementById('assignMediaModal').dataset.originalFilePath = filePath;
    
    // Najpierw wygeneruj checkboxy grup
    updateMediaGroupSelect();
    
    // Następnie zaznacz grupy, do których należą te media
    // Pobierz ID wszystkich media dla tego pliku
    const mediaIds = mediaItems.map(m => m.id);
    
    // Przeszukaj wszystkie grupy i zaznacz te, które zawierają którekolwiek z tych media
    setTimeout(() => {
        mediaGroups.forEach(group => {
            // Sprawdź czy grupa zawiera którekolwiek z tych media
            // (To jest uproszczenie - w pełnej implementacji trzeba by sprawdzić media_group_items)
            // Na razie zostawiamy niezaznaczone, bo przy edycji nie zachowujemy grup
        });
    }, 100);
    
    document.getElementById('assignMediaModal').classList.add('active');
}

async function removeMediaFile(filePath) {
    if (!confirm('Czy na pewno usunąć to media ze wszystkich scen?')) return;

    try {
        const mediaItems = assignedMedia.filter(m => m.file_path === filePath);
        
        for (const item of mediaItems) {
            await fetch(`/api/episodes/${currentEpisodeId}/media/${item.id}`, {
                method: 'DELETE'
            });
        }
        
        await loadAssignedMedia();
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd usuwania');
    }
}

async function removeMediaFromEpisode(mediaId) {
    if (!confirm('Czy na pewno chcesz usunąć to przypisanie?')) return;

    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/media/${mediaId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            await loadAssignedMedia();
        } else {
            alert('Błąd usuwania media');
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

// ===== UTILITIES =====
function formatDuration(seconds) {
    const minutes = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${minutes}:${secs.toString().padStart(2, '0')}`;
}

// ===== MEDIA GROUPS =====
async function loadMediaGroups() {
    if (!currentEpisodeId) return;
    
    try {
        const response = await fetch(`/api/media-groups?episode_id=${currentEpisodeId}`);
        mediaGroups = await response.json();
        renderMediaGroups();
    } catch (error) {
        console.error('Błąd ładowania grup mediów:', error);
    }
}

function renderMediaGroups() {
    const containerMedia = document.getElementById('mediaGroupsListMedia');
    const containerReportaze = document.getElementById('mediaGroupsListReportaze');
    
    if (mediaGroups.length === 0) {
        containerMedia.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak grup MEDIA</div>';
        containerReportaze.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak grup REPORTAŻE</div>';
        return;
    }

    // Rozdziel grupy według scen
    const mediaGroups_MEDIA = mediaGroups.filter(g => g.scene && g.scene.name === 'MEDIA');
    const mediaGroups_REPORTAZE = mediaGroups.filter(g => g.scene && g.scene.name === 'REPORTAZE');

    // Renderuj grupy MEDIA
    if (mediaGroups_MEDIA.length === 0) {
        containerMedia.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak grup MEDIA</div>';
    } else {
        containerMedia.innerHTML = mediaGroups_MEDIA.map(group => {
            const isActive = group.is_current || false;
            const activeClass = isActive ? 'active' : '';
            
            return `
                <div class="media-group-card ${activeClass}" data-group-id="${group.id}" onclick="openManageMediaGroupModal(${group.id})">
                    <div class="media-group-header">
                        <div>
                            <div class="media-group-name">${group.name}</div>
                            <div class="media-group-count">${group.description || 'Brak opisu'}</div>
                        </div>
                        ${isActive ? '<span class="badge badge-success">AKTYWNA</span>' : ''}
                    </div>
                </div>
            `;
        }).join('');
    }

    // Renderuj grupy REPORTAZE
    if (mediaGroups_REPORTAZE.length === 0) {
        containerReportaze.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak grup REPORTAŻE</div>';
    } else {
        containerReportaze.innerHTML = mediaGroups_REPORTAZE.map(group => {
            const isActive = group.is_current || false;
            const activeClass = isActive ? 'active' : '';
            
            return `
                <div class="media-group-card ${activeClass}" data-group-id="${group.id}" onclick="openManageMediaGroupModal(${group.id})">
                    <div class="media-group-header">
                        <div>
                            <div class="media-group-name">${group.name}</div>
                            <div class="media-group-count">${group.description || 'Brak opisu'}</div>
                        </div>
                        ${isActive ? '<span class="badge badge-success">AKTYWNA</span>' : ''}
                    </div>
                </div>
            `;
        }).join('');
    }
}

function openAddMediaGroupModal() {
    document.getElementById('addMediaGroupForm').reset();
    document.getElementById('groupSceneMedia').checked = false;
    document.getElementById('groupSceneReportaze').checked = false;
    document.getElementById('groupSceneError').style.display = 'none';
    document.getElementById('addMediaGroupModal').classList.add('active');
}

function closeAddMediaGroupModal() {
    document.getElementById('addMediaGroupModal').classList.remove('active');
}

async function createMediaGroup() {
    if (!currentEpisodeId) {
        alert('Najpierw zapisz odcinek');
        return;
    }

    // Walidacja scen
    const mediaChecked = document.getElementById('groupSceneMedia').checked;
    const reportazeChecked = document.getElementById('groupSceneReportaze').checked;
    
    if (!mediaChecked && !reportazeChecked) {
        document.getElementById('groupSceneError').style.display = 'block';
        return;
    }

    // Utwórz grupę dla każdej wybranej sceny
    try {
        const selectedScenes = [];
        if (mediaChecked) selectedScenes.push({name: 'MEDIA', id: mediaSceneId});
        if (reportazeChecked) selectedScenes.push({name: 'REPORTAZE', id: reportazeSceneId});
        
        for (const scene of selectedScenes) {
            const data = {
                episode_id: currentEpisodeId,
                scene_id: scene.id,
                name: document.getElementById('mediaGroupName').value + (selectedScenes.length > 1 ? ` (${scene.name})` : ''),
                description: document.getElementById('mediaGroupDescription').value
            };
            
            const response = await fetch('/api/media-groups', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(data)
            });

            if (!response.ok) {
                const error = await response.text();
                alert('Błąd dodawania grupy: ' + error);
                return;
            }
        }
        
        closeAddMediaGroupModal();
        await loadMediaGroups();
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

async function openManageMediaGroupModal(groupId) {
    currentMediaGroup = mediaGroups.find(g => g.id === groupId);
    if (!currentMediaGroup) return;

    document.getElementById('currentMediaGroupId').value = groupId;
    document.getElementById('manageMediaGroupTitle').textContent = currentMediaGroup.name;
    document.getElementById('mediaGroupInfo').textContent = currentMediaGroup.description || 'Brak opisu';

    // Sprawdź czy istnieją grupy o tej samej nazwie w innych scenach
    const sceneName = currentMediaGroup.scene ? currentMediaGroup.scene.name : '';
    const sameNameGroups = mediaGroups.filter(g => 
        g.name === currentMediaGroup.name && 
        g.episode_id === currentMediaGroup.episode_id
    );
    
    // Zaznacz checkboxy dla wszystkich scen gdzie istnieje ta grupa
    let hasMedia = sameNameGroups.some(g => g.scene?.name === 'MEDIA');
    let hasReportaze = sameNameGroups.some(g => g.scene?.name === 'REPORTAZE');
    
    document.getElementById('manageGroupSceneMedia').checked = hasMedia;
    document.getElementById('manageGroupSceneReportaze').checked = hasReportaze;

    // Załaduj media w grupie
    await loadGroupMediaItems(groupId);

    document.getElementById('manageMediaGroupModal').classList.add('active');
}

function closeManageMediaGroupModal() {
    document.getElementById('manageMediaGroupModal').classList.remove('active');
    currentMediaGroup = null;
}

async function updateGroupScenes() {
    const groupId = parseInt(document.getElementById('currentMediaGroupId').value);
    if (!groupId || !currentMediaGroup) return;
    
    const mediaChecked = document.getElementById('manageGroupSceneMedia').checked;
    const reportazeChecked = document.getElementById('manageGroupSceneReportaze').checked;
    
    // Wymaga przynajmniej jednej sceny
    if (!mediaChecked && !reportazeChecked) {
        alert('Grupa musi być przypisana do przynajmniej jednej sceny');
        // Przywróć poprzedni stan
        const currentSceneName = currentMediaGroup.scene ? currentMediaGroup.scene.name : '';
        document.getElementById('manageGroupSceneMedia').checked = (currentSceneName === 'MEDIA');
        document.getElementById('manageGroupSceneReportaze').checked = (currentSceneName === 'REPORTAZE');
        return;
    }
    
    const currentSceneName = currentMediaGroup.scene ? currentMediaGroup.scene.name : '';
    const currentSceneWasMedia = (currentSceneName === 'MEDIA');
    const currentSceneWasReportaze = (currentSceneName === 'REPORTAZE');
    
    // Sprawdź co się zmieniło
    const nowWantsMedia = mediaChecked;
    const nowWantsReportaze = reportazeChecked;
    
    // Jeśli nic się nie zmieniło, wyjdź
    if (currentSceneWasMedia === nowWantsMedia && currentSceneWasReportaze === nowWantsReportaze) {
        return;
    }
    
    try {
        // Znajdź wszystkie grupy o tej samej nazwie w tym odcinku
        const sameNameGroups = mediaGroups.filter(g => 
            g.name === currentMediaGroup.name && 
            g.episode_id === currentMediaGroup.episode_id
        );
        
        const existingMediaGroup = sameNameGroups.find(g => g.scene?.name === 'MEDIA');
        const existingReportazeGroup = sameNameGroups.find(g => g.scene?.name === 'REPORTAZE');
        
        // Przypadek 1: Chcemy dodać MEDIA (nie było wcześniej)
        if (nowWantsMedia && !existingMediaGroup) {
            const mediaScene = window.mediaScenes.find(s => s.name === 'MEDIA');
            if (mediaScene) {
                const response = await fetch('/api/media-groups', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({
                        episode_id: currentMediaGroup.episode_id,
                        scene_id: mediaScene.id,
                        name: currentMediaGroup.name,
                        description: currentMediaGroup.description
                    })
                });
                
                if (!response.ok) {
                    alert('Błąd tworzenia grupy MEDIA');
                    document.getElementById('manageGroupSceneMedia').checked = false;
                    return;
                }
            }
        }
        
        // Przypadek 2: Chcemy dodać REPORTAZE (nie było wcześniej)
        if (nowWantsReportaze && !existingReportazeGroup) {
            const reportazeScene = window.mediaScenes.find(s => s.name === 'REPORTAZE');
            if (reportazeScene) {
                const response = await fetch('/api/media-groups', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({
                        episode_id: currentMediaGroup.episode_id,
                        scene_id: reportazeScene.id,
                        name: currentMediaGroup.name,
                        description: currentMediaGroup.description
                    })
                });
                
                if (!response.ok) {
                    alert('Błąd tworzenia grupy REPORTAZE');
                    document.getElementById('manageGroupSceneReportaze').checked = false;
                    return;
                }
            }
        }
        
        // Przypadek 3: Chcemy usunąć MEDIA (było wcześniej)
        if (!nowWantsMedia && existingMediaGroup) {
            const response = await fetch(`/api/media-groups/${existingMediaGroup.id}`, {
                method: 'DELETE'
            });
            
            if (!response.ok) {
                alert('Błąd usuwania grupy MEDIA');
                document.getElementById('manageGroupSceneMedia').checked = true;
                return;
            }
        }
        
        // Przypadek 4: Chcemy usunąć REPORTAZE (było wcześniej)
        if (!nowWantsReportaze && existingReportazeGroup) {
            const response = await fetch(`/api/media-groups/${existingReportazeGroup.id}`, {
                method: 'DELETE'
            });
            
            if (!response.ok) {
                alert('Błąd usuwania grupy REPORTAZE');
                document.getElementById('manageGroupSceneReportaze').checked = true;
                return;
            }
        }
        
        // Odśwież listę grup
        await loadMediaGroups();
        
        // Zamknij modal tylko jeśli odznaczono bieżącą scenę
        if ((currentSceneWasMedia && !nowWantsMedia) || (currentSceneWasReportaze && !nowWantsReportaze)) {
            // Sprawdź czy pozostała jakaś scena
            if (!nowWantsMedia && !nowWantsReportaze) {
                closeManageMediaGroupModal();
            } else {
                // Jeśli pozostała inna scena, zaktualizuj currentMediaGroup
                const remainingGroups = mediaGroups.filter(g => 
                    g.name === currentMediaGroup.name && 
                    g.episode_id === currentMediaGroup.episode_id
                );
                if (remainingGroups.length > 0) {
                    currentMediaGroup = remainingGroups[0];
                    document.getElementById('currentMediaGroupId').value = currentMediaGroup.id;
                }
            }
        }
        
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
        // Przywróć poprzedni stan checkboxów
        document.getElementById('manageGroupSceneMedia').checked = currentSceneWasMedia;
        document.getElementById('manageGroupSceneReportaze').checked = currentSceneWasReportaze;
    }
}

async function loadGroupMediaItems(groupId) {
    const container = document.getElementById('groupMediaList');
    
    try {
        const response = await fetch(`/api/media-groups/${groupId}/items`);
        const items = await response.json();

        if (items.length === 0) {
            container.innerHTML = '<div style="text-align: center; color: #666; padding: 20px;">Brak mediów w grupie</div>';
            return;
        }

        // Sortuj po kolejności
        items.sort((a, b) => a.order - b.order);

        container.innerHTML = items.map(item => {
            const media = item.episode_media;
            return `
                <div class="group-media-item" data-item-id="${item.id}">
                    <div style="flex: 1;">
                        <strong>${media.title}</strong>
                        <div style="font-size: 10px; color: #666;">
                            ${media.scene ? media.scene.name : ''} • 
                            ${media.duration ? formatDuration(media.duration) : ''}
                        </div>
                    </div>
                    <div class="group-media-order">
                        <button class="btn btn-danger btn-icon" 
                                onclick="removeMediaFromCurrentGroup(${groupId}, ${media.id})"
                                title="Usuń z grupy">×</button>
                    </div>
                </div>
            `;
        }).join('');
        
        // Inicjalizuj Sortable dla drag & drop
        initGroupMediaItemsSortable();
    } catch (error) {
        console.error('Błąd ładowania mediów grupy:', error);
        container.innerHTML = '<div style="text-align: center; color: #f00; padding: 20px;">Błąd ładowania</div>';
    }
}

async function addMediaToCurrentGroup(mediaId) {
    const groupId = document.getElementById('currentMediaGroupId').value;
    
    // Pobierz aktualną liczbę mediów w grupie dla kolejności
    let maxOrder = 0;
    try {
        const response = await fetch(`/api/media-groups/${groupId}/items`);
        const items = await response.json();
        if (items.length > 0) {
            maxOrder = Math.max(...items.map(i => i.order));
        }
    } catch (error) {
        console.error('Błąd:', error);
    }

    try {
        const response = await fetch(`/api/media-groups/${groupId}/media/${mediaId}`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ order: maxOrder + 1 })
        });

        if (response.ok) {
            await loadGroupMediaItems(groupId);
        } else {
            const error = await response.text();
            alert('Błąd dodawania: ' + error);
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

async function removeMediaFromCurrentGroup(groupId, mediaId) {
    if (!confirm('Czy na pewno chcesz usunąć to media z grupy?')) return;

    try {
        const response = await fetch(`/api/media-groups/${groupId}/media/${mediaId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            await loadGroupMediaItems(groupId);
        } else {
            alert('Błąd usuwania');
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

async function deleteMediaGroup() {
    const groupId = document.getElementById('currentMediaGroupId').value;
    
    if (!confirm('Czy na pewno chcesz usunąć tę grupę? Ta operacja jest nieodwracalna.')) return;

    try {
        const response = await fetch(`/api/media-groups/${groupId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            closeManageMediaGroupModal();
            await loadMediaGroups();
        } else {
            alert('Błąd usuwania grupy');
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

async function setMediaGroupAsCurrent() {
    const groupId = document.getElementById('currentMediaGroupId').value;
    
    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/media-groups/${groupId}/set-current`, {
            method: 'POST'
        });

        if (response.ok) {
            alert('Grupa ustawiona jako aktywna (wczytana do źródła List)');
            closeManageMediaGroupModal();
            await loadMediaGroups();
        } else {
            alert('Błąd ustawiania grupy');
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}
// ===== DRAG & DROP - SORTABLE =====
function initAssignedMediaSortable() {
    const container = document.getElementById('assignedMediaList');
    if (!container || assignedMedia.length === 0) return;
    
    new Sortable(container, {
        animation: 150,
        ghostClass: 'sortable-ghost',
        dragClass: 'sortable-drag',
        onEnd: async function(evt) {
            const itemId = evt.item.dataset.mediaId;
            const newOrder = evt.newIndex;
            
            try {
                await fetch(`/api/episodes/${currentEpisodeId}/media/${itemId}/reorder`, {
                    method: 'PUT',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({ order: newOrder })
                });
                await loadAssignedMedia();
            } catch (error) {
                console.error('Błąd aktualizacji kolejności:', error);
                alert('Błąd aktualizacji kolejności');
            }
        }
    });
}

function initMediaGroupsSortable() {
    const container = document.getElementById('mediaGroupsList');
    if (!container || mediaGroups.length === 0) return;
    
    new Sortable(container, {
        animation: 150,
        ghostClass: 'sortable-ghost',
        dragClass: 'sortable-drag',
        onEnd: async function(evt) {
            const groupId = evt.item.dataset.groupId;
            const newOrder = evt.newIndex;
            
            try {
                await fetch(`/api/media-groups/${groupId}/reorder`, {
                    method: 'PUT',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({ order: newOrder })
                });
                await loadMediaGroups();
            } catch (error) {
                console.error('Błąd aktualizacji kolejności:', error);
                alert('Błąd aktualizacji kolejności');
            }
        }
    });
}

function initGroupMediaItemsSortable() {
    const container = document.getElementById('groupMediaList');
    if (!container || !currentMediaGroup) return;
    
    const groupId = currentMediaGroup.id;
    
    new Sortable(container, {
        animation: 150,
        ghostClass: 'sortable-ghost',
        dragClass: 'sortable-drag',
        onEnd: async function(evt) {
            const itemId = parseInt(evt.item.getAttribute('data-item-id'));
            const newOrder = evt.newIndex;
            
            if (!itemId) {
                console.error('Brak itemId');
                return;
            }
            
            try {
                const response = await fetch(`/api/media-groups/${groupId}/items/${itemId}/reorder`, {
                    method: 'PUT',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({ order: newOrder })
                });
                
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${await response.text()}`);
                }
                
                await loadGroupMediaItems(groupId);
            } catch (error) {
                console.error('Błąd aktualizacji kolejności:', error);
                alert('Błąd aktualizacji kolejności: ' + error.message);
            }
        }
    });
}