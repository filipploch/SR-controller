// vlc_group_modal.js - Modal wyboru grupy dla VLC Video Source (Media2/Reportaze2)

let currentVLCModalSourceName = null;
let vlcModalData = null;

// Otwórz modal wyboru grupy
async function openVLCGroupModal(sourceName, sceneName) {
    currentVLCModalSourceName = sourceName;

    if (!currentEpisodeId) {
        alert('Brak aktualnego odcinka');
        return;
    }

    try {
        const response = await fetch(`/api/episodes/${currentEpisodeId}/sources/${sourceName}/groups-list`);
        const data = await response.json();

        vlcModalData = data;

        // Renderuj modal
        renderVLCGroupModal(data, sourceName, sceneName);

        // Pokaż modal
        document.getElementById('vlc-group-modal-overlay').classList.add('active');
    } catch (error) {
        console.error('Błąd ładowania grup:', error);
        alert('Nie udało się załadować list grup');
    }
}

// Renderuj zawartość modalu
function renderVLCGroupModal(data, sourceName, sceneName) {
    const modalTitle = document.getElementById('vlc-group-modal-title');
    modalTitle.textContent = `Wybierz grupę dla ${sourceName}`;

    const modalBody = document.getElementById('vlc-group-modal-body');
    modalBody.innerHTML = '';

    if (!data.groups || data.groups.length === 0) {
        modalBody.innerHTML = '<p class="no-groups">Brak grup z co najmniej 2 plikami</p>';
        return;
    }

    // Renderuj grupy
    data.groups.forEach(group => {
        const groupDiv = document.createElement('div');
        groupDiv.className = 'vlc-group-item';

        if (group.is_current) {
            groupDiv.classList.add('active');
        }

        // Ikona i nazwa
        const icon = group.is_system ? '📁' : '📂';
        const systemLabel = group.is_system ? ' (systemowa)' : '';
        
        groupDiv.innerHTML = `
            <div class="vlc-group-info">
                <span class="vlc-group-icon">${icon}</span>
                <span class="vlc-group-name">${group.name}${systemLabel}</span>
            </div>
            <div class="vlc-group-count">${group.file_count} ${group.file_count === 1 ? 'plik' : 'pliki/plików'}</div>
        `;

        // Dwuklik - przypisz grupę
        groupDiv.ondblclick = () => {
            assignGroupToSource(group.id, group.name);
        };

        modalBody.appendChild(groupDiv);
    });
}

// Przypisz grupę do źródła
async function assignGroupToSource(groupId, groupName) {
    if (!currentEpisodeId || !currentVLCModalSourceName) {
        alert('Błąd: brak danych odcinka lub źródła');
        return;
    }

    try {
        const response = await fetch(
            `/api/episodes/${currentEpisodeId}/sources/${currentVLCModalSourceName}/assign-group`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    group_id: groupId
                })
            }
        );

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        const result = await response.json();
        console.log('Przypisano grupę:', result);

        // Zapisz nazwę źródła przed zamknięciem modalu
        const sourceName = currentVLCModalSourceName;

        // Zamknij modal
        closeVLCGroupModal();

        // Zaktualizuj tekst przycisku lokalnie (broadcast też zaktualizuje)
        updateSourceButtonText(sourceName, groupName);

    } catch (error) {
        console.error('Błąd przypisywania grupy:', error);
        alert('Nie udało się przypisać grupy: ' + error.message);
    }
}

// Zamknij modal
function closeVLCGroupModal() {
    document.getElementById('vlc-group-modal-overlay').classList.remove('active');
    currentVLCModalSourceName = null;
    vlcModalData = null;
}

// Event listeners
document.addEventListener('DOMContentLoaded', () => {
    // Zamknij modal po kliknięciu na overlay
    const overlay = document.getElementById('vlc-group-modal-overlay');
    if (overlay) {
        overlay.addEventListener('click', (e) => {
            if (e.target.id === 'vlc-group-modal-overlay') {
                closeVLCGroupModal();
            }
        });
    }

    // Zamknij modal po kliknięciu X
    const closeBtn = document.getElementById('vlc-group-modal-close');
    if (closeBtn) {
        closeBtn.addEventListener('click', closeVLCGroupModal);
    }
});