// ===== MODAL WYBORU MEDIÓW =====

// Pobierz aktualny odcinek przy ładowaniu
async function loadCurrentEpisode() {
    try {
        const response = await fetch('/api/episodes?current=true');
        const episodes = await response.json();
        if (episodes && episodes.length > 0) {
            currentEpisodeId = episodes[0].id;
            return episodes[0];
        }
    } catch (error) {
        console.error('Błąd ładowania aktualnego odcinka:', error);
    }
    return null;
}

// Otwórz modal wyboru mediów
async function openMediaModal(sourceName, sceneName) {
    if (!currentEpisodeId) {
        alert('Brak aktualnego odcinka');
        return;
    }

    currentModalSourceName = sourceName;

    // Pobierz dane dla modalu
    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/sources/${sourceName}/media-list`);
        modalData = await response.json();

        renderMediaModal(modalData, sourceName, sceneName);
        document.getElementById('mediaModalOverlay').classList.add('active');
    } catch (error) {
        console.error('Błąd ładowania danych modalu:', error);
        alert('Nie udało się załadować listy mediów');
    }
}

// Renderuj modal
function renderMediaModal(data, sourceName, sceneName) {
    const modalBody = document.getElementById('mediaModalBody');
    const modalMediaTitle = document.getElementById('mediaModalTitle');

    modalMediaTitle.textContent = `Wybierz media dla ${sourceName} (${sceneName})`;
    modalBody.innerHTML = '';

    if (!data.groups || data.groups.length === 0) {
        modalBody.innerHTML = '<div class="no-media">Brak grup mediów dla tego odcinka</div>';
        return;
    }

    // Znajdź grupę z aktualnym plikiem i przenieś ją na początek
    let currentGroupIndex = -1;
    if (data.current_media_id) {
        currentGroupIndex = data.groups.findIndex(g => g.is_current);
    }

    let orderedGroups = [...data.groups];
    if (currentGroupIndex > 0) {
        const currentGroup = orderedGroups.splice(currentGroupIndex, 1)[0];
        orderedGroups.unshift(currentGroup);
    }

    // Renderuj grupy
    orderedGroups.forEach((group, index) => {
        const groupDiv = document.createElement('div');
        groupDiv.className = 'media-group' + (group.is_system ? ' system' : '');
        groupDiv.dataset.groupId = group.id;

        // Rozwiń pierwszą grupę (z aktualnym plikiem) lub pierwszą grupę jeśli brak aktualnego
        if (index === 0) {
            groupDiv.classList.add('expanded');
        }

        // Header grupy
        const headerDiv = document.createElement('div');
        headerDiv.className = 'media-group-header';
        headerDiv.innerHTML = `
            <span class="media-group-name">${group.name}</span>
            <span class="media-group-toggle">▼</span>
        `;
        headerDiv.onclick = () => toggleMediaGroup(group.id);

        // Lista mediów
        const itemsDiv = document.createElement('div');
        itemsDiv.className = 'media-items-list';

        if (group.media_items && group.media_items.length > 0) {
            group.media_items.forEach(media => {
                const itemDiv = document.createElement('div');
                itemDiv.className = 'media-item';
                itemDiv.dataset.mediaId = media.id;
                itemDiv.textContent = media.title;

                // Oznacz aktualny plik
                if (data.current_media_id && media.id === data.current_media_id) {
                    itemDiv.classList.add('active');
                }

                // Dwuklik - przypisz plik
                itemDiv.ondblclick = () => assignMediaToSource(media.id, media.title);

                itemsDiv.appendChild(itemDiv);
            });
        } else {
            itemsDiv.innerHTML = '<div class="no-media">Brak mediów w tej grupie</div>';
        }

        groupDiv.appendChild(headerDiv);
        groupDiv.appendChild(itemsDiv);
        modalBody.appendChild(groupDiv);
    });
}

// Przełącz grupę (zwiń/rozwiń)
function toggleMediaGroup(groupId) {
    const allGroups = document.querySelectorAll('.media-group');
    const clickedGroup = document.querySelector(`.media-group[data-group-id="${groupId}"]`);

    if (!clickedGroup) return;

    const wasExpanded = clickedGroup.classList.contains('expanded');

    if (wasExpanded) {
        // Kliknięto na rozwiniętą grupę - po prostu ją zwiń
        clickedGroup.classList.remove('expanded');
    } else {
        // Kliknięto na zwiniętą grupę
        // 1. Zwiń wszystkie grupy
        allGroups.forEach(g => g.classList.remove('expanded'));

        // 2. Przenieś klikniętą grupę pod pierwszą (jeśli nie jest pierwsza)
        const modalBody = document.getElementById('mediaModalBody');
        const firstGroup = modalBody.firstChild;

        if (clickedGroup !== firstGroup) {
            modalBody.insertBefore(clickedGroup, firstGroup.nextSibling);
        }

        // 3. Rozwiń klikniętą grupę
        clickedGroup.classList.add('expanded');
    }
}

// Przypisz plik do źródła
async function assignMediaToSource(mediaId, mediaTitle) {
    if (!currentEpisodeId || !currentModalSourceName) {
        alert('Błąd - brak danych o odcinku lub źródle');
        return;
    }

    try {
        const response = await fetch(
            `/api/episodes/${currentEpisodeId}/sources/${currentModalSourceName}/assign-media`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    media_id: mediaId
                })
            }
        );

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        const result = await response.json();
        console.log('Przypisano media:', result);

        // Zapisz nazwę źródła przed zamknięciem modalu
        const sourceName = currentModalSourceName;

        // Zamknij modal
        closeMediaModal();

        // Zaktualizuj tekst przycisku lokalnie (broadcast też zaktualizuje)
        updateSourceButtonText(sourceName, mediaTitle);

    } catch (error) {
        console.error('Błąd przypisywania media:', error);
        alert('Nie udało się przypisać pliku: ' + error.message);
    }
}

// Zamknij modal
function closeMediaModal() {
    document.getElementById('mediaModalOverlay').classList.remove('active');
    currentModalSourceName = null;
    modalData = null;
}

// Zaktualizuj tekst przycisku źródła
function updateSourceButtonText(sourceName, newText) {
    console.log(`[updateSourceButtonText] Szukam przycisku: ${sourceName}, nowy tekst: ${newText}`);
    const button = document.querySelector(`[data-source-name="${sourceName}"]`);
    if (button) {
        console.log(`[updateSourceButtonText] Znaleziono przycisk, aktualizuję tekst`);
        button.textContent = newText;
    } else {
        console.warn(`[updateSourceButtonText] NIE znaleziono przycisku dla: ${sourceName}`);
    }
}

// Nasłuchuj na broadcast przypisania media
socket.on('source_media_assigned', (data) => {
    console.log('Otrzymano broadcast przypisania media:', data);

    // Sprawdź czy dotyczy aktualnego odcinka
    if (currentEpisodeId && data.episode_id !== currentEpisodeId) {
        console.log('Ignoruję - inny odcinek');
        return;
    }

    // Zaktualizuj przycisk
    updateSourceButtonText(data.source_name, data.title);
});

// Nasłuchuj na broadcast przypisania grupy (VLC Video Source)
socket.on('source_group_assigned', (data) => {
    console.log('Otrzymano broadcast przypisania grupy:', data);

    // Sprawdź czy dotyczy aktualnego odcinka
    if (currentEpisodeId && data.episode_id !== currentEpisodeId) {
        console.log('Ignoruję - inny odcinek');
        return;
    }

    // Zaktualizuj przycisk
    updateSourceButtonText(data.source_name, data.name);
});

// Nasłuchuj na broadcast przypisania typu kamery
socket.on('source_camera_assigned', (data) => {
    console.log('Otrzymano broadcast przypisania kamery:', data);

    // Sprawdź czy dotyczy aktualnego odcinka
    if (currentEpisodeId && data.episode_id !== currentEpisodeId) {
        console.log('Ignoruję - inny odcinek');
        return;
    }

    // Zaktualizuj przycisk kamery
    updateCameraButtonState(
        data.source_name,
        data.camera_type_name || data.source_name,
        data.is_disabled || false
    );
});

// Załaduj wszystkie przypisania przy starcie
async function loadAllSourceAssignments() {
    if (!currentEpisodeId) {
        console.log('Brak aktualnego odcinka - pomijam ładowanie przypisań');
        return;
    }

    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/source-assignments`);
        const assignments = await response.json();

        console.log('Załadowano przypisania źródeł:', assignments);

        // Zaktualizuj przyciski
        for (const [sourceName, assignment] of Object.entries(assignments)) {
            if (assignment.type === 'camera') {
                // Kamera - użyj updateCameraButtonState
                updateCameraButtonState(
                    sourceName,
                    assignment.button_text,
                    assignment.is_disabled || false
                );
            } else {
                // Media/Group - użyj updateSourceButtonText
                updateSourceButtonText(sourceName, assignment.button_text);
            }
        }
    } catch (error) {
        console.error('Błąd ładowania przypisań źródeł:', error);
    }
}

// Automatyczne przypisanie Media1 i Reportaze1 jeśli nie mają przypisań
async function autoAssignMediaSources() {
    if (!currentEpisodeId) {
        console.log('Brak aktualnego odcinka - pomijam automatyczne przypisanie');
        return;
    }

    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/auto-assign-media-sources`, {
            method: 'POST'
        });
        const result = await response.json();

        console.log('Wynik automatycznego przypisania Media1/Reportaze1:', result);

        // Zaktualizuj przyciski dla przypisanych źródeł
        if (result.Media1 && result.Media1.assigned) {
            updateSourceButtonText('Media1', result.Media1.title);
        }
        if (result.Reportaze1 && result.Reportaze1.assigned) {
            updateSourceButtonText('Reportaze1', result.Reportaze1.title);
        }
    } catch (error) {
        console.error('Błąd automatycznego przypisania Media1/Reportaze1:', error);
    }
}

// Automatyczne przypisanie Media2 i Reportaze2 jeśli nie mają przypisań
async function autoAssignVLCSources() {
    if (!currentEpisodeId) {
        console.log('Brak aktualnego odcinka - pomijam automatyczne przypisanie VLC');
        return;
    }

    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/auto-assign-vlc-sources`, {
            method: 'POST'
        });
        const result = await response.json();

        console.log('Wynik automatycznego przypisania Media2/Reportaze2:', result);

        // Zaktualizuj przyciski dla przypisanych źródeł
        if (result.Media2 && result.Media2.assigned) {
            updateSourceButtonText('Media2', result.Media2.name);
        }
        if (result.Reportaze2 && result.Reportaze2.assigned) {
            updateSourceButtonText('Reportaze2', result.Reportaze2.name);
        }
    } catch (error) {
        console.error('Błąd automatycznego przypisania:', error);
    }
}

// Automatyczne przypisanie typów kamer (Kamera1-4)
// async function autoAssignCameraTypes() {
//     if (!currentEpisodeId) {
//         console.log('Brak aktualnego odcinka - pomijam automatyczne przypisanie kamer');
//         return;
//     }

//     try {
//         const response = await fetch(`/api/episodes/${currentEpisodeId}/auto-assign-camera-types`, {
//             method: 'POST'
//         });
//         const result = await response.json();

//         console.log('Wynik automatycznego przypisania kamer:', result);

//         // Zaktualizuj przyciski dla przypisanych kamer
//         ['Kamera1', 'Kamera2', 'Kamera3', 'Kamera4'].forEach(sourceName => {
//             if (result[sourceName] && result[sourceName].assigned) {
//                 updateCameraButtonState(
//                     sourceName,
//                     result[sourceName].camera_type_name,
//                     false // nie wyłączona
//                 );
//             }
//         });
//     } catch (error) {
//         console.error('Błąd automatycznego przypisania kamer:', error);
//     }
// }

async function loadCameraAssignments() {
    const response = await fetch(`/api/episodes/${currentEpisodeId}/camera-assignments`);  // ← GET, read-only
    const assignments = await response.json();

    for (const [sourceName, assignment] of Object.entries(assignments)) {
        updateCameraButtonState(
            sourceName,
            assignment.camera_type_name,
            assignment.is_disabled || false
        );
    }
}

// Aktualizuj stan przycisku kamery (tekst + disabled)
function updateCameraButtonState(sourceName, cameraTypeName, isDisabled) {
    const button = document.querySelector(`[data-source-name="${sourceName}"]`);
    if (button) {
        if (isDisabled) {
            // Wyłączona
            button.textContent = sourceName; // "Kamera1"
            button.disabled = true;
            button.classList.add('camera-disabled');
        } else {
            // Włączona
            button.textContent = cameraTypeName || sourceName; // "Centralna" lub "Kamera1"
            button.disabled = false;
            button.classList.remove('camera-disabled');
        }
    }
}


// Event listeners
document.addEventListener('DOMContentLoaded', () => {
    // Zamknij modal po kliknięciu na overlay
    document.getElementById('mediaModalOverlay').addEventListener('click', (e) => {
        if (e.target.id === 'mediaModalOverlay') {
            closeMediaModal();
        }
    });

    // Zamknij modal po kliknięciu X
    document.getElementById('mediaModalClose').addEventListener('click', closeMediaModal);
});

// Załaduj aktualny odcinek przy połączeniu
socket.on('connect', () => {
    loadCurrentEpisode().then(() => {
        // Automatycznie przypisz Media1/Reportaze1 i Media2/Reportaze2
        // (renderSources wywoła loadAllSourceAssignments() po wyrenderowaniu przycisków)
        autoAssignMediaSources().then(() => {
            autoAssignVLCSources().then(() => {
                // autoAssignCameraTypes();
                loadCameraAssignments()
            });
        });
    });
});