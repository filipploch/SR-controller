// microphone_modal.js - Modal przypisania osoby do mikrofonu

// Otwórz modal wyboru osoby dla mikrofonu
function openMicrophoneAssignModal(sourceName, sceneName) {
    console.log(`Opening microphone modal for ${sourceName}`);
    
    fetch(`/api/episodes/${currentEpisodeId}/sources/${sourceName}/microphone-people-list`)
        .then(response => response.json())
        .then(data => {
            console.log(data);
            renderMicrophoneModal(data, sourceName);
        })
        .catch(error => {
            console.error('Error loading people list:', error);
            alert('Nie udało się załadować listy osób');
        });
}

// Renderuj modal z listą Staff + Guests
function renderMicrophoneModal(data, sourceName) {
    const overlay = document.getElementById('microphone-modal-overlay');
    const title = document.getElementById('microphone-modal-title');
    const body = document.getElementById('microphone-modal-body');
    
    title.textContent = `Przypisz osobę: ${sourceName}`;
    body.innerHTML = '';
    
    // Sekcja Staff
    if (data.staff && data.staff.length > 0) {
        const staffHeader = document.createElement('div');
        staffHeader.className = 'microphone-section-header';
        staffHeader.textContent = 'PROWADZĄCY';
        body.appendChild(staffHeader);
        
        data.staff.forEach(person => {
            const item = createPersonItem(person, data.current_person_id, data.current_person_type, sourceName);
            body.appendChild(item);
        });
    }
    
    // Sekcja Guests
    if (data.guests && data.guests.length > 0) {
        const guestHeader = document.createElement('div');
        guestHeader.className = 'microphone-section-header';
        guestHeader.textContent = 'GOŚCIE';
        body.appendChild(guestHeader);
        
        data.guests.forEach(person => {
            const item = createPersonItem(person, data.current_person_id, data.current_person_type, sourceName);
            body.appendChild(item);
        });
    }
    
    // Opcja "Usuń przypisanie"
    const unassignItem = document.createElement('div');
    unassignItem.className = 'microphone-person-item unassign-option';
    unassignItem.innerHTML = `
        <span>Usuń przypisanie</span>
        <span class="microphone-count"></span>
    `;
    unassignItem.addEventListener('dblclick', () => {
        unassignMicrophonePerson(sourceName);
        closeMicrophoneModal();
    });
    body.appendChild(unassignItem);
    
    // Pokaż modal
    overlay.classList.add('active');
}

// Utwórz element osoby
function createPersonItem(person, currentPersonId, currentPersonType, sourceName) {
    const item = document.createElement('div');
    item.className = 'microphone-person-item';
    
    const isCurrent = person.is_current;
    const microphones = person.assigned_microphones || [];
    
    if (isCurrent) {
        item.classList.add('active');
    }
    
    item.innerHTML = `
        <span>${person.last_name} ${person.first_name}</span>
        <span class="microphone-count">${microphones.length > 0 ? microphones.join(', ') : ''}</span>
    `;
    
    // Dwuklik = przypisz
    item.addEventListener('dblclick', () => {
        assignMicrophonePerson(person.id, person.type, person.full_name, sourceName);
        closeMicrophoneModal();
    });
    
    return item;
}

// Przypisz osobę do mikrofonu
function assignMicrophonePerson(personId, personType, personName, sourceName) {
    fetch(`/api/episodes/${currentEpisodeId}/sources/${sourceName}/assign-microphone-person`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            person_id: personId,
            person_type: personType
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            console.log(`Assigned ${personName} to ${sourceName}`);
            // Aktualizuj lokalnie
            updateMicrophoneButtonText(sourceName, personName);
        } else {
            alert('Nie udało się przypisać osoby: ' + (data.error || 'unknown error'));
        }
    })
    .catch(error => {
        console.error('Error assigning person:', error);
        alert('Błąd przypisania osoby');
    });
}

// Usuń przypisanie osoby z mikrofonu
function unassignMicrophonePerson(sourceName) {
    fetch(`/api/episodes/${currentEpisodeId}/sources/${sourceName}/assign-microphone-person`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            person_id: null,
            person_type: ""
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            console.log(`Unassigned person from ${sourceName}`);
            // Przywróć oryginalną nazwę
            updateMicrophoneButtonText(sourceName, sourceName);
        } else {
            alert('Nie udało się usunąć przypisania: ' + (data.error || 'unknown error'));
        }
    })
    .catch(error => {
        console.error('Error unassigning person:', error);
        alert('Błąd usuwania przypisania');
    });
}

// Aktualizuj tekst przycisku mikrofonu
function updateMicrophoneButtonText(sourceName, newText) {
    const button = document.querySelector(`button[data-source-name="${sourceName}"]`);
    if (button) {
        button.textContent = newText;
        console.log(`Updated button ${sourceName} to: ${newText}`);
    }
}

// Zamknij modal
function closeMicrophoneModal() {
    const overlay = document.getElementById('microphone-modal-overlay');
    overlay.classList.remove('active');
}

// Listener - zamknięcie modalu przez X
document.addEventListener('DOMContentLoaded', () => {
    const closeBtn = document.getElementById('microphone-modal-close');
    if (closeBtn) {
        closeBtn.addEventListener('click', closeMicrophoneModal);
    }
    
    // Zamknięcie przez kliknięcie poza modalem
    const overlay = document.getElementById('microphone-modal-overlay');
    if (overlay) {
        overlay.addEventListener('click', (e) => {
            if (e.target === overlay) {
                closeMicrophoneModal();
            }
        });
    }
});

// WebSocket listener - broadcast przypisania mikrofonu
socket.on('source_microphone_assigned', (data) => {
    console.log('Otrzymano broadcast przypisania mikrofonu:', data);
    updateMicrophoneButtonText(data.source_name, data.person_name);
});