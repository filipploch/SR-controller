// ===== INITIALIZATION =====
document.addEventListener('DOMContentLoaded', () => {
    loadSeasonsForSelect();
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
function switchEpisodeTab(tabName, sourceElement) {
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
async function loadSeasonsForSelect() {
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

async function openCreateEpisodeModal() {
    document.getElementById('modalEpisodeTitle').textContent = 'Nowy Odcinek';
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
    switchEpisodeTab('data');
    
    document.getElementById('episodeModal').classList.add('active');
}

function openEditModal(id) {
    const episode = episodes.find(e => e.id === id);
    if (!episode) return;

    currentEpisodeId = id;
    document.getElementById('modalEpisodeTitle').textContent = 'Edycja Odcinka';
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
    switchEpisodeTab('data');
    
    document.getElementById('episodeModal').classList.add('active');
}

async function closeEpisodeModal() {
    document.getElementById('episodeModal').classList.remove('active');
    
    // Przywróć currentEpisodeId do aktualnego odcinka
    // (nie ustawiaj null, bo to czyści przypisania w kontrolerze!)
    try {
        const response = await fetch('/api/episodes?current=true');
        const currentEpisodes = await response.json();
        if (currentEpisodes && currentEpisodes.length > 0) {
            currentEpisodeId = currentEpisodes[0].id;
            
            // Odśwież przypisania w kontrolerze
            if (typeof loadAllSourceAssignments === 'function') {
                await loadAllSourceAssignments();
            }
        } else {
            currentEpisodeId = null;
        }
        console.log("currentEpisodeId: ", currentEpisodeId);
    } catch (error) {
        console.error('Błąd przywracania aktualnego odcinka:', error);
        currentEpisodeId = null;
    }
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
            closeEpisodeModal();
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
        if (!response.ok) return;
        
        const scenes = await response.json();
        console.log('Sceny mediów:', scenes);
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

// async function loadMediaFiles() {
//     if (!currentEpisodeId) return;
    
//     try {
//         const response = await fetch(`/api/episodes/${currentEpisodeId}/media/files`);
//         availableMediaFiles = await response.json();
//         renderMediaFiles();
//     } catch (error) {
//         console.error('Błąd ładowania plików:', error);
//     }
// }

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
    // Ustaw dane pliku
    document.getElementById('mediaFilePath').value = path;
    document.getElementById('mediaFileDuration').value = duration || 0;
    document.getElementById('mediaFileName').textContent = name;
    document.getElementById('mediaTitle').value = name.replace(/\.[^/.]+$/, ''); // Usuń rozszerzenie
    
    // Wyczyść tryb edycji (jeśli był aktywny)
    const modal = document.getElementById('assignMediaModal');
    delete modal.dataset.editMode;
    delete modal.dataset.editMediaId;
    
    // Załaduj grupy do select
    loadGroupsToSelect();
    
    // Otwórz modal
    document.getElementById('assignMediaModal').classList.add('active');
}

function loadGroupsToSelect() {
    const groupsSelect = document.getElementById('mediaGroupsSelect');
    if (!groupsSelect) {
        console.error('Element mediaGroupsSelect nie istnieje!');
        return;
    }
    
    groupsSelect.innerHTML = '';
    
    if (!mediaGroups || mediaGroups.length === 0) {
        groupsSelect.innerHTML = '<option value="" disabled>Brak dostępnych grup</option>';
        return;
    }
    
    // Rozdziel na systemowe i użytkownika
    const systemGroups = mediaGroups.filter(g => g.is_system);
    const userGroups = mediaGroups.filter(g => !g.is_system);
    
    // Dodaj grupy systemowe
    if (systemGroups.length > 0) {
        const optgroup = document.createElement('optgroup');
        optgroup.label = 'Grupy systemowe';
        systemGroups.forEach(group => {
            const option = document.createElement('option');
            option.value = group.id;
            option.textContent = group.name;
            optgroup.appendChild(option);
        });
        groupsSelect.appendChild(optgroup);
    }
    
    // Dodaj grupy użytkownika
    if (userGroups.length > 0) {
        const optgroup = document.createElement('optgroup');
        optgroup.label = 'Grupy użytkownika';
        userGroups.forEach(group => {
            const option = document.createElement('option');
            option.value = group.id;
            option.textContent = group.name;
            optgroup.appendChild(option);
        });
        groupsSelect.appendChild(optgroup);
    }
}

function openAssignMediaModal(filePath, duration, fileName) {
    // Jeśli przekazano parametry, wypełnij formularz
    if (filePath) {
        document.getElementById('mediaFilePath').value = filePath;
        document.getElementById('mediaFileDuration').value = duration || 0;
        document.getElementById('mediaFileName').textContent = fileName || filePath;
        document.getElementById('mediaTitle').value = fileName ? fileName.replace(/\.[^/.]+$/, '') : '';
    }
    
    document.getElementById('mediaDescription').value = '';
    
    // Załaduj listę Staff (autorów)
    const staffSelect = document.getElementById('mediaStaff');
    staffSelect.innerHTML = '<option value="">Brak</option>';
    
    if (assignedStaff && assignedStaff.length > 0) {
        assignedStaff.forEach(assignment => {
            if (assignment.staff) {
                const option = document.createElement('option');
                option.value = assignment.id;
                option.textContent = `${assignment.staff.first_name} ${assignment.staff.last_name}`;
                staffSelect.appendChild(option);
            }
        });
    }
    
    // Wyczyść tryb edycji
    const modal = document.getElementById('assignMediaModal');
    delete modal.dataset.editMode;
    delete modal.dataset.editMediaId;
    
    // Załaduj grupy
    loadGroupsToSelect();
    
    document.getElementById('assignMediaModal').classList.add('active');
}

function closeAssignMediaModal() {
    const modal = document.getElementById('assignMediaModal');
    // modal.style.display = 'none';
    // LUB jeśli używasz classList:
    modal.classList.remove('active');
    
    // Wyczyść tryb edycji
    delete modal.dataset.editMode;
    delete modal.dataset.editMediaId;
    
    // Resetuj formularz
    document.getElementById('assignMediaForm').reset();
}

// Zaktualizowana funkcja assignMedia aby obsługiwała edycję
async function assignMedia() {
    const modal = document.getElementById('assignMediaModal');
    const editMode = modal.dataset.editMode === 'true';
    const editMediaId = editMode ? parseInt(modal.dataset.editMediaId) : null;
    
    const filePath = document.getElementById('mediaFilePath').value;
    const duration = parseInt(document.getElementById('mediaFileDuration').value) || 0;
    const title = document.getElementById('mediaTitle').value;
    const description = document.getElementById('mediaDescription').value;
    const staffId = document.getElementById('mediaStaff').value;
    
    // Pobierz wybrane grupy
    const groupsSelect = document.getElementById('mediaGroupsSelect');
    const selectedGroupIds = Array.from(groupsSelect.selectedOptions).map(opt => parseInt(opt.value));
    
    if (selectedGroupIds.length === 0) {
        alert('Wybierz przynajmniej jedną grupę!');
        return;
    }
    
    try {
        if (editMode && editMediaId) {
            // EDYCJA ISTNIEJĄCEGO MEDIA
            
            // 1. Zaktualizuj podstawowe dane media
            const updateData = {
                title: title,
                description: description,
                episode_staff_id: staffId ? parseInt(staffId) : null
            };
            
            const updateResponse = await fetch(`/api/episodes/${currentEpisodeId}/media/${editMediaId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(updateData)
            });
            
            if (!updateResponse.ok) {
                const error = await updateResponse.text();
                throw new Error(error);
            }
            
            // 2. Pobierz aktualne przypisania do grup
            const currentMedia = assignedMedia.find(m => m.id === editMediaId);
            const currentGroupIds = currentMedia && currentMedia.media_groups 
                ? currentMedia.media_groups.map(mg => mg.media_group_id) 
                : [];
            
            // 3. Usuń z grup które nie są już wybrane
            const groupsToRemove = currentGroupIds.filter(id => !selectedGroupIds.includes(id));
            for (const groupId of groupsToRemove) {
                await fetch(`/api/media-groups/${groupId}/media/${editMediaId}`, {
                    method: 'DELETE'
                });
            }
            
            // 4. Dodaj do nowych grup
            const groupsToAdd = selectedGroupIds.filter(id => !currentGroupIds.includes(id));
            for (const groupId of groupsToAdd) {
                const assignmentResponse = await fetch(`/api/media-groups/${groupId}/items`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        episode_media_id: editMediaId
                    })
                });
                
                if (!assignmentResponse.ok && assignmentResponse.status !== 409) {
                    // 409 = już istnieje, ignoruj ten błąd
                    console.error(`Nie udało się przypisać do grupy ${groupId}`);
                }
            }
            
            alert('Media zaktualizowane pomyślnie!');
            
        } else {
            // TWORZENIE NOWEGO MEDIA
            
            // 1. Utwórz media
            const mediaData = {
                title: title,
                description: description,
                file_path: filePath,
                duration: duration,
                episode_staff_id: staffId ? parseInt(staffId) : null
            };
            
            const mediaResponse = await fetch(`/api/episodes/${currentEpisodeId}/media`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(mediaData)
            });
            
            if (!mediaResponse.ok) {
                const error = await mediaResponse.text();
                throw new Error(error);
            }
            
            const createdMedia = await mediaResponse.json();
            
            // 2. Przypisz do wybranych grup
            for (const groupId of selectedGroupIds) {
                const assignmentResponse = await fetch(`/api/media-groups/${groupId}/items`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        episode_media_id: createdMedia.id
                    })
                });
                
                if (!assignmentResponse.ok && assignmentResponse.status !== 409) {
                    console.error(`Nie udało się przypisać do grupy ${groupId}`);
                }
            }
            
            alert('Media przypisane pomyślnie!');
        }
        
        // Wyczyść tryb edycji
        delete modal.dataset.editMode;
        delete modal.dataset.editMediaId;
        
        closeAssignMediaModal();
        // await loadEpisodeDetails(currentEpisodeId);
        // ----->
        await Promise.all([
            loadAssignedMedia(),
            loadMediaGroups()
        ]);
        renderAssignedMedia();
        renderMediaGroups();
        await loadMediaFiles();
        // <-----
        switchMediaSubTab('assigned');
        
    } catch (error) {
        console.error('Błąd zapisywania media:', error);
        alert('Błąd: ' + error.message);
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
        container.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak przypisanych mediów</div>';
        return;
    }

    // Renderuj wszystkie media w jednej liście
    const html = assignedMedia.map(media => {
        const authorName = media.episode_staff && media.episode_staff.staff ? 
            `${media.episode_staff.staff.first_name} ${media.episode_staff.staff.last_name}` : 
            'Brak';
        
        // Wyświetl grupy do których należy media
        const groupNames = media.media_groups && media.media_groups.length > 0
            ? media.media_groups.map(mg => mg.media_group.name).join(', ')
            : 'Brak grup';
        
        return `
            <div class="assigned-media-item">
                <div class="media-item-details">
                    <div class="media-item-title">${media.title}</div>
                    <div class="media-item-meta">
                        Autor: ${authorName}<br>
                        Grupy: ${groupNames}<br>
                        ${media.description ? `Opis: ${media.description}<br>` : ''}
                        Plik: ${media.file_path || 'Brak'}<br>
                        ${media.duration ? `Czas: ${formatDuration(media.duration)}<br>` : ''}
                    </div>
                </div>
                <div class="list-item-actions" style="pointer-events: auto;">
                    <button class="btn btn-primary btn-icon" onclick="editMediaAssignment(${media.id})" title="Edytuj">✎</button>
                    <button class="btn btn-danger btn-icon" onclick="removeMedia(${media.id})" title="Usuń">×</button>
                </div>
            </div>
        `;
    }).join('');
    
    container.innerHTML = html;
}

function editMediaAssignment(mediaId) {
    // Znajdź media po ID
    const media = assignedMedia.find(m => m.id === mediaId);
    if (!media) {
        console.error('Nie znaleziono media o ID:', mediaId);
        return;
    }
    
    // Wypełnij formularz
    document.getElementById('mediaFilePath').value = media.file_path;
    document.getElementById('mediaFileDuration').value = media.duration || 0;
    document.getElementById('mediaFileName').textContent = media.file_path ? media.file_path.split('/').pop() : '';
    document.getElementById('mediaTitle').value = media.title;
    document.getElementById('mediaDescription').value = media.description || '';
    document.getElementById('mediaStaff').value = media.episode_staff_id || '';
    
    // Zapisz ID media jako tryb edycji
    document.getElementById('assignMediaModal').dataset.editMode = 'true';
    document.getElementById('assignMediaModal').dataset.editMediaId = mediaId;
    
    // Załaduj grupy do select
    const groupsSelect = document.getElementById('mediaGroupsSelect');
    groupsSelect.innerHTML = '';
    
    if (mediaGroups && mediaGroups.length > 0) {
        // Grupy systemowe
        const systemGroups = mediaGroups.filter(g => g.is_system);
        const userGroups = mediaGroups.filter(g => !g.is_system);
        
        if (systemGroups.length > 0) {
            const optgroup = document.createElement('optgroup');
            optgroup.label = 'Grupy systemowe';
            systemGroups.forEach(group => {
                const option = document.createElement('option');
                option.value = group.id;
                option.textContent = group.name;
                optgroup.appendChild(option);
            });
            groupsSelect.appendChild(optgroup);
        }
        
        if (userGroups.length > 0) {
            const optgroup = document.createElement('optgroup');
            optgroup.label = 'Grupy użytkownika';
            userGroups.forEach(group => {
                const option = document.createElement('option');
                option.value = group.id;
                option.textContent = group.name;
                optgroup.appendChild(option);
            });
            groupsSelect.appendChild(optgroup);
        }
    }
    
    // Zaznacz grupy do których należy media
    if (media.media_groups && media.media_groups.length > 0) {
        const mediaGroupIds = media.media_groups.map(mg => mg.media_group_id);
        
        // Zaznacz opcje w select
        Array.from(groupsSelect.options).forEach(option => {
            if (mediaGroupIds.includes(parseInt(option.value))) {
                option.selected = true;
            }
        });
    }
    
    document.getElementById('assignMediaModal').classList.add('active');
}

async function removeMedia(mediaId) {
    if (!confirm('Czy na pewno chcesz usunąć to media? Zostanie usunięte ze wszystkich grup.')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/media/${mediaId}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            throw new Error('Nie udało się usunąć media');
        }
        
        await Promise.all([
            loadAssignedMedia(),
            loadMediaGroups()
        ]);
        renderAssignedMedia();
        renderMediaGroups();
        await loadMediaFiles();
        alert('Media usunięte pomyślnie!');
        
    } catch (error) {
        console.error('Błąd usuwania media:', error);
        alert('Błąd: ' + error.message);
    }
}

async function loadGroupMedia(groupId) {
    try {
        const response = await fetch(`/api/media-groups/${groupId}/items`);
        if (!response.ok) {
            throw new Error('Nie udało się załadować mediów w grupie');
        }
        
        const items = await response.json();
        const container = document.getElementById('groupMediaList');
        
        if (items.length === 0) {
            container.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak mediów w grupie</div>';
            return;
        }
        
        const html = items.map(item => {
            const media = item.episode_media;
            return `
                <div class="group-media-item">
                    <div style="flex: 1;">
                        <div style="font-weight: 500; font-size: 12px;">${media.title}</div>
                        <div style="font-size: 10px; color: #888;">
                            ${media.file_path || 'Brak pliku'}
                            ${media.duration ? ` • ${formatDuration(media.duration)}` : ''}
                        </div>
                    </div>
                    <button class="btn btn-danger btn-icon btn-small" 
                            onclick="removeMediaFromGroup(${groupId}, ${media.id})" 
                            title="Usuń z grupy">×</button>
                </div>
            `;
        }).join('');
        
        container.innerHTML = html;
        
    } catch (error) {
        console.error('Błąd ładowania mediów grupy:', error);
        document.getElementById('groupMediaList').innerHTML = 
            '<div style="text-align: center; padding: 20px; color: #f00; font-size: 11px;">Błąd ładowania</div>';
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
function formatDuration(milliseconds) {
    const totalSeconds = milliseconds / 1000;
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = Math.floor(totalSeconds % 60);
    const ms = Math.floor((milliseconds % 1000));
    
    return `${minutes}:${seconds.toString().padStart(2, '0')}.${ms.toString().padStart(3, '0')}`;
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
    const systemContainer = document.getElementById('systemGroupsList');
    const userContainer = document.getElementById('userGroupsList');
    
    if (mediaGroups.length === 0) {
        systemContainer.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Ładowanie...</div>';
        userContainer.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak grup użytkownika</div>';
        return;
    }
    
    // Rozdziel na systemowe i użytkownika
    const systemGroups = mediaGroups.filter(g => g.is_system);
    const userGroups = mediaGroups.filter(g => !g.is_system);
    
    // Funkcja renderująca pojedynczą grupę
    const renderGroup = (group, isSystem) => {
        const icon = group.name === 'MEDIA' ? '📺' : 
                     group.name === 'REPORTAZE' ? '🎬' : '📁';
        
        return `
            <div class="media-group-item" onclick="openManageMediaGroupModal(${group.id})">
                <div style="display: flex; align-items: center; gap: 8px;">
                    <span style="font-size: 18px;">${icon}</span>
                    <div>
                        <div style="font-weight: 500; font-size: 13px;">${group.name}</div>
                        ${group.description ? `<div style="font-size: 10px; color: #888;">${group.description}</div>` : ''}
                    </div>
                </div>
                ${isSystem ? '<span class="badge" style="background: #555;">Systemowa</span>' : ''}
            </div>
        `;
    };
    
    // Renderuj grupy systemowe
    if (systemGroups.length === 0) {
        systemContainer.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak grup systemowych</div>';
    } else {
        systemContainer.innerHTML = systemGroups.map(g => renderGroup(g, true)).join('');
    }
    
    // Renderuj grupy użytkownika
    if (userGroups.length === 0) {
        userContainer.innerHTML = '<div style="text-align: center; padding: 20px; color: #666; font-size: 11px;">Brak grup użytkownika</div>';
    } else {
        userContainer.innerHTML = userGroups.map(g => renderGroup(g, false)).join('');
    }
}

function openAddMediaGroupModal() {
    document.getElementById('addMediaGroupForm').reset();
    
    // Sprawdź czy checkboxy istnieją (stary kod)
    const sceneMediaCheckbox = document.getElementById('groupSceneMedia');
    const sceneReportazeCheckbox = document.getElementById('groupSceneReportaze');
    const sceneError = document.getElementById('groupSceneError');
    
    if (sceneMediaCheckbox) sceneMediaCheckbox.checked = false;
    if (sceneReportazeCheckbox) sceneReportazeCheckbox.checked = false;
    if (sceneError) sceneError.style.display = 'none';
    
    document.getElementById('addMediaGroupModal').classList.add('active');
}

function closeAddMediaGroupModal() {
    document.getElementById('addMediaGroupModal').classList.remove('active');
}

async function createMediaGroup() {
    const name = document.getElementById('groupName').value;
    const description = document.getElementById('groupDescription').value;
    
    // USUŃ: checkboxy scen, nie wysyłaj scene_id
    
    try {
        const response = await fetch('/api/media-groups', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                episode_id: currentEpisodeId,
                name: name,
                description: description
            })
        });
        
        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }
        
        closeAddMediaGroupModal();
        await Promise.all([
            loadAssignedMedia(),
            loadMediaGroups()
        ]);
        renderAssignedMedia();
        renderMediaGroups();
        await loadMediaFiles();
        switchMediaSubTab('groups');
        alert('Grupa utworzona pomyślnie!');
        
    } catch (error) {
        console.error('Błąd tworzenia grupy:', error);
        alert('Błąd: ' + error.message);
    }
}

async function openManageMediaGroupModal(groupId) {
    currentMediaGroup = mediaGroups.find(g => g.id === groupId);
    if (!currentMediaGroup) return;
    
    document.getElementById('manageGroupId').value = groupId;
    document.getElementById('manageGroupTitle').textContent = `Zarządzanie Grupą: ${currentMediaGroup.name}`;
    document.getElementById('manageGroupName').value = currentMediaGroup.name;
    document.getElementById('manageGroupDescription').value = currentMediaGroup.description || '';
    
    // Ukryj przycisk Usuń dla grup systemowych
    const deleteBtn = document.getElementById('deleteGroupBtn');
    if (currentMediaGroup.is_system) {
        deleteBtn.style.display = 'none';
        // Zablokuj edycję nazwy dla grup systemowych
        document.getElementById('manageGroupName').disabled = true;
    } else {
        deleteBtn.style.display = 'inline-block';
        document.getElementById('manageGroupName').disabled = false;
    }
    
    // USUŃ: aktualizację checkboxów scen
    
    // Załaduj media w grupie
    await loadGroupMedia(groupId);
    
    document.getElementById('manageMediaGroupModal').classList.add('active');
}

async function updateMediaGroup() {
    const groupId = document.getElementById('manageGroupId').value;
    const name = document.getElementById('manageGroupName').value;
    const description = document.getElementById('manageGroupDescription').value;
    
    // USUŃ: pobieranie i wysyłanie scen
    
    try {
        const response = await fetch(`/api/media-groups/${groupId}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                name: name,
                description: description
            })
        });
        
        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }
        
        closeManageMediaGroupModal();
        await Promise.all([
            loadAssignedMedia(),
            loadMediaGroups()
        ]);
        renderAssignedMedia();
        renderMediaGroups();
        await loadMediaFiles();
        alert('Grupa zaktualizowana pomyślnie!');
        
    } catch (error) {
        console.error('Błąd aktualizacji grupy:', error);
        alert('Błąd: ' + error.message);
    }
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
                                onclick="removeMediaFromGroup(${groupId}, ${media.id})"
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

async function removeMediaFromGroup(groupId, mediaId) {
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