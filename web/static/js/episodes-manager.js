// Global state
let episodes = [];
let seasons = [];
let allStaff = [];
let allGuests = [];
let staffTypes = [];
let guestTypes = [];
let sources = [];
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
function switchTab(tabName) {
    // Update tab buttons
    document.querySelectorAll('.modal-tab').forEach(tab => {
        tab.classList.remove('active');
    });
    event.target.classList.add('active');

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
        const date = episode.episode_date ? new Date(episode.episode_date).toLocaleDateString('pl-PL') : '-';
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

function openCreateModal() {
    document.getElementById('modalTitle').textContent = 'Nowy Odcinek';
    document.getElementById('episodeForm').reset();
    document.getElementById('episodeId').value = '';
    currentEpisodeId = null;
    assignedStaff = [];
    assignedGuests = [];
    assignedMedia = [];
    
    // Switch to first tab
    switchTab('data');
    document.querySelector('.modal-tab').click();
    
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
    document.querySelector('.modal-tab').click();
    
    document.getElementById('episodeModal').classList.add('active');
}

function closeModal() {
    document.getElementById('episodeModal').classList.remove('active');
    currentEpisodeId = null;
}

async function saveEpisode() {
    const id = document.getElementById('episodeId').value;
    const dateValue = document.getElementById('episodeDate').value;
    
    const data = {
        season_id: parseInt(document.getElementById('episodeSeason').value),
        episode_number: parseInt(document.getElementById('episodeNumber').value),
        season_episode: parseInt(document.getElementById('seasonEpisode').value),
        title: document.getElementById('episodeTitle').value,
        episode_date: dateValue ? new Date(dateValue).toISOString() : new Date().toISOString(),
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
        
        updateMediaSourceSelect();
        updateMediaStaffSelect();
    } catch (error) {
        console.error('Błąd ładowania scen:', error);
    }
}

function updateMediaSourceSelect() {
    const select = document.getElementById('mediaSource');
    if (!window.mediaScenes) {
        select.innerHTML = '<option value="">Ładowanie...</option>';
        return;
    }
    
    select.innerHTML = '<option value="">Wybierz...</option>' +
        window.mediaScenes.map(scene => 
            `<option value="${scene.id}">${scene.name}</option>`
        ).join('');
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
    const select = document.getElementById('mediaGroupSelect');
    if (!select) return; // Element nie istnieje jeśli formularz nie jest otwarty
    select.innerHTML = '<option value="">Nie przypisuj do grupy</option>' +
        mediaGroups.map(group => 
            `<option value="${group.id}">${group.name}</option>`
        ).join('');
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

    container.innerHTML = availableMediaFiles.map(file => `
        <div class="media-file-card" onclick="selectMediaFile('${file.path}', '${file.name}', ${file.duration})">
            <div class="media-file-name">${file.name}</div>
            <div class="media-file-info">
                Typ: ${file.type}<br>
                ${file.duration ? `Czas: ${formatDuration(file.duration)}` : ''}
            </div>
        </div>
    `).join('');
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

    const filePath = document.getElementById('mediaFilePath').value;
    const staffId = document.getElementById('mediaStaff').value;
    const groupId = document.getElementById('mediaGroupSelect').value;
    
    const data = {
        scene_id: parseInt(document.getElementById('mediaSource').value),
        title: document.getElementById('mediaTitle').value,
        description: document.getElementById('mediaDescription').value,
        file_path: filePath,
        duration: parseInt(document.getElementById('mediaFileDuration').value),
        episode_staff_id: staffId ? parseInt(staffId) : null
    };

    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/media`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(data)
        });

        if (response.ok) {
            const newMedia = await response.json();
            
            // Jeśli wybrano grupę, dodaj media do grupy
            if (groupId) {
                try {
                    await fetch(`/api/media-groups/${groupId}/items`, {
                        method: 'POST',
                        headers: {'Content-Type': 'application/json'},
                        body: JSON.stringify({
                            episode_media_id: newMedia.id
                        })
                    });
                } catch (error) {
                    console.error('Błąd dodawania do grupy:', error);
                    // Nie przerywamy - media zostało przypisane, tylko nie dodane do grupy
                }
            }
            
            closeAssignMediaModal();
            await loadAssignedMedia();
        } else {
            const error = await response.text();
            alert('Błąd przypisywania media: ' + error);
        }
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
    const container = document.getElementById('assignedMediaList');
    
    if (assignedMedia.length === 0) {
        container.innerHTML = '<div style="text-align: center; padding: 20px; color: #666;">Brak przypisanych mediów</div>';
        return;
    }

    container.innerHTML = assignedMedia.map(media => {
        const sceneName = media.scene ? media.scene.name : 'Brak';
        const authorName = media.episode_staff && media.episode_staff.staff ? 
            `${media.episode_staff.staff.first_name} ${media.episode_staff.staff.last_name}` : 
            'Brak';
        const currentBadge = media.is_current ? 
            '<span class="badge badge-success">WCZYTANY</span>' : '';
        
        return `
            <div class="assigned-media-item">
                <div class="media-item-details">
                    <div class="media-item-title">${media.title} ${currentBadge}</div>
                    <div class="media-item-meta">
                        Scena: ${sceneName}<br>
                        Autor: ${authorName}<br>
                        ${media.description ? `Opis: ${media.description}<br>` : ''}
                        Plik: ${media.file_path || media.url || 'Brak'}<br>
                        ${media.duration ? `Czas: ${formatDuration(media.duration)}<br>` : ''}
                    </div>
                </div>
                <div class="list-item-actions">
                    ${!media.is_current ? `<button class="btn btn-success btn-icon" onclick="setCurrentMedia(${media.id})" title="Wczytaj do źródła Single">⬆</button>` : ''}
                    <button class="btn btn-danger btn-icon" onclick="removeMediaFromEpisode(${media.id})">×</button>
                </div>
            </div>
        `;
    }).join('');
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

async function setCurrentMedia(mediaId) {
    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/media/${mediaId}/set-current`, {
            method: 'POST'
        });

        if (response.ok) {
            await loadAssignedMedia();
        } else {
            alert('Błąd ustawiania media');
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
    try {
        const response = await fetch('/api/media-groups');
        mediaGroups = await response.json();
        renderMediaGroups();
    } catch (error) {
        console.error('Błąd ładowania grup mediów:', error);
    }
}

function renderMediaGroups() {
    const container = document.getElementById('mediaGroupsList');
    
    if (mediaGroups.length === 0) {
        container.innerHTML = '<div style="text-align: center; padding: 20px; color: #666;">Brak grup mediów</div>';
        return;
    }

    container.innerHTML = mediaGroups.map(group => {
        // Policz ile mediów jest w grupie (potrzebujemy zapytania do API)
        const isActive = false; // TODO: sprawdź czy grupa jest aktywna
        const activeClass = isActive ? 'active' : '';
        
        return `
            <div class="media-group-card ${activeClass}" onclick="openManageMediaGroupModal(${group.id})">
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

function openAddMediaGroupModal() {
    document.getElementById('addMediaGroupForm').reset();
    document.getElementById('addMediaGroupModal').classList.add('active');
}

function closeAddMediaGroupModal() {
    document.getElementById('addMediaGroupModal').classList.remove('active');
}

async function createMediaGroup() {
    const data = {
        name: document.getElementById('mediaGroupName').value,
        description: document.getElementById('mediaGroupDescription').value
    };

    try {
        const response = await fetch('/api/media-groups', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(data)
        });

        if (response.ok) {
            closeAddMediaGroupModal();
            await loadMediaGroups();
        } else {
            const error = await response.text();
            alert('Błąd dodawania grupy: ' + error);
        }
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

    // Załaduj dostępne media i media w grupie
    await loadAvailableMediaForGroup(groupId);
    await loadGroupMediaItems(groupId);

    document.getElementById('manageMediaGroupModal').classList.add('active');
}

function closeManageMediaGroupModal() {
    document.getElementById('manageMediaGroupModal').classList.remove('active');
    currentMediaGroup = null;
}

async function loadAvailableMediaForGroup(groupId) {
    const container = document.getElementById('availableMediaForGroup');
    
    // Pobierz media już w grupie
    let groupMediaIds = [];
    try {
        const response = await fetch(`/api/media-groups/${groupId}/items`);
        const items = await response.json();
        groupMediaIds = items.map(item => item.episode_media_id);
    } catch (error) {
        console.error('Błąd ładowania mediów grupy:', error);
    }

    // Filtruj dostępne media (przypisane do odcinka ale nie w grupie)
    const available = assignedMedia.filter(m => !groupMediaIds.includes(m.id));

    if (available.length === 0) {
        container.innerHTML = '<div style="text-align: center; color: #666; padding: 20px;">Wszystkie media są już w grupach</div>';
        return;
    }

    container.innerHTML = available.map(media => `
        <div class="available-media-item">
            <div>
                <strong>${media.title}</strong>
                <div style="font-size: 10px; color: #666;">${media.scene ? media.scene.name : ''}</div>
            </div>
            <button class="btn btn-success btn-icon" onclick="addMediaToCurrentGroup(${media.id})" title="Dodaj do grupy">+</button>
        </div>
    `).join('');
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
        items.sort((a, b) => a.episode_order - b.episode_order);

        container.innerHTML = items.map(item => {
            const media = item.episode_media;
            return `
                <div class="group-media-item">
                    <div style="flex: 1;">
                        <strong>${media.title}</strong>
                        <div style="font-size: 10px; color: #666;">
                            ${media.scene ? media.scene.name : ''} • 
                            ${media.duration ? formatDuration(media.duration) : ''}
                        </div>
                    </div>
                    <div class="group-media-order">
                        <input type="number" value="${item.episode_order}" min="1" 
                               onchange="updateMediaOrder(${groupId}, ${item.id}, this.value)"
                               title="Kolejność">
                        <button class="btn btn-danger btn-icon" 
                                onclick="removeMediaFromCurrentGroup(${groupId}, ${media.id})"
                                title="Usuń z grupy">×</button>
                    </div>
                </div>
            `;
        }).join('');
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
            maxOrder = Math.max(...items.map(i => i.episode_order));
        }
    } catch (error) {
        console.error('Błąd:', error);
    }

    try {
        const response = await fetch(`/api/media-groups/${groupId}/media/${mediaId}`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ episode_order: maxOrder + 1 })
        });

        if (response.ok) {
            await loadAvailableMediaForGroup(groupId);
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
            await loadAvailableMediaForGroup(groupId);
            await loadGroupMediaItems(groupId);
        } else {
            alert('Błąd usuwania');
        }
    } catch (error) {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    }
}

async function updateMediaOrder(groupId, itemId, newOrder) {
    // TODO: Endpoint do aktualizacji kolejności
    console.log('Update order:', groupId, itemId, newOrder);
    // Możemy to zrobić przez usunięcie i dodanie z nową kolejnością
    // lub dodać dedykowany endpoint
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