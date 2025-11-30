// camera_type_modal.js - Modal wyboru typu kamery

let currentCameraModalSourceName = null;
let currentCameraModalSceneName = null;

// Otwórz modal wyboru typu kamery
async function openCameraTypeModal(sourceName, sceneName) {
    currentCameraModalSourceName = sourceName;
    currentCameraModalSceneName = sceneName;

    if (!currentEpisodeId) {
        alert('Brak aktualnego odcinka');
        return;
    }

    try {
        const response = await fetch(
            `/api/episodes/${currentEpisodeId}/sources/${sourceName}/camera-types-list`
        );
        const data = await response.json();

        renderCameraTypeModal(data, sourceName);
        document.getElementById('camera-type-modal-overlay').classList.add('active');
    } catch (error) {
        console.error('Błąd ładowania typów kamer:', error);
        alert('Nie udało się załadować listy typów kamer');
    }
}

// Renderuj modal z listą typów
function renderCameraTypeModal(data, sourceName) {
    const modalTitle = document.getElementById('camera-type-modal-title');
    const modalBody = document.getElementById('camera-type-modal-body');

    modalTitle.textContent = `Wybierz typ kamery - ${sourceName}`;
    modalBody.innerHTML = '';

    if (!data.camera_types || data.camera_types.length === 0) {
        modalBody.innerHTML = '<p class="no-types">Brak typów kamer</p>';
        return;
    }

    // Renderuj typy kamer
    data.camera_types.forEach(type => {
        const typeDiv = document.createElement('div');
        typeDiv.className = 'camera-type-item';

        if (type.is_current) {
            typeDiv.classList.add('active');
        }

        if (type.is_assigned && !type.is_current) {
            typeDiv.classList.add('disabled');
        }

        // Ikona i nazwa
        const icon = type.is_system ? '📹' : '🎥';
        const systemLabel = type.is_system ? ' (systemowy)' : '';
        const assignedLabel = type.is_assigned && !type.is_current 
            ? ` - przypisany do ${type.assigned_to}` 
            : '';

        typeDiv.innerHTML = `
            <div class="camera-type-info">
                <span class="camera-type-icon">${icon}</span>
                <span class="camera-type-name">${type.name}${systemLabel}${assignedLabel}</span>
            </div>
            <div class="camera-type-order">#${type.order}</div>
        `;

        // Dwuklik - przypisz typ (jeśli nie jest już przypisany do innej kamery)
        if (!type.is_assigned || type.is_current) {
            typeDiv.ondblclick = () => {
                assignCameraTypeToSource(type.id, type.name);
            };
        } else {
            typeDiv.style.cursor = 'not-allowed';
            typeDiv.title = `Ten typ jest już przypisany do ${type.assigned_to}`;
        }

        modalBody.appendChild(typeDiv);
    });

    // Opcja "Wyłącz kamerę" na dole
    const disableDiv = document.createElement('div');
    disableDiv.className = 'camera-type-item disable-option';
    disableDiv.innerHTML = `
        <div class="camera-type-info">
            <span class="camera-type-icon">❌</span>
            <span class="camera-type-name">Wyłącz kamerę</span>
        </div>
    `;

    disableDiv.ondblclick = () => {
        disableCameraSource();
    };

    modalBody.appendChild(disableDiv);
}

// Przypisz typ kamery
async function assignCameraTypeToSource(cameraTypeId, cameraTypeName) {
    if (!currentEpisodeId || !currentCameraModalSourceName) {
        return;
    }

    try {
        const response = await fetch(
            `/api/episodes/${currentEpisodeId}/sources/${currentCameraModalSourceName}/assign-camera-type`,
            {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    camera_type_id: cameraTypeId
                })
            }
        );

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        closeCameraTypeModal();

        // Lokalnie zaktualizuj przycisk
        updateCameraButtonState(currentCameraModalSourceName, cameraTypeName, false);

    } catch (error) {
        console.error('Błąd przypisywania typu kamery:', error);
        alert('Nie udało się przypisać typu: ' + error.message);
    }
}

// Wyłącz kamerę
async function disableCameraSource() {
    if (!currentEpisodeId || !currentCameraModalSourceName) {
        return;
    }

    try {
        const response = await fetch(
            `/api/episodes/${currentEpisodeId}/sources/${currentCameraModalSourceName}/assign-camera-type`,
            {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    camera_type_id: null  // NULL = wyłącz
                })
            }
        );

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        closeCameraTypeModal();

        // Lokalnie zaktualizuj przycisk - wyłączona
        updateCameraButtonState(currentCameraModalSourceName, currentCameraModalSourceName, true);

    } catch (error) {
        console.error('Błąd wyłączania kamery:', error);
        alert('Nie udało się wyłączyć kamery: ' + error.message);
    }
}

// Zamknij modal
function closeCameraTypeModal() {
    document.getElementById('camera-type-modal-overlay').classList.remove('active');
    currentCameraModalSourceName = null;
    currentCameraModalSceneName = null;
}

// Event listeners
document.addEventListener('DOMContentLoaded', () => {
    // Zamknij modal po kliknięciu na overlay
    const overlay = document.getElementById('camera-type-modal-overlay');
    if (overlay) {
        overlay.addEventListener('click', (e) => {
            if (e.target === overlay) {
                closeCameraTypeModal();
            }
        });
    }

    // Zamknij modal po kliknięciu na X
    const closeBtn = document.getElementById('camera-type-modal-close');
    if (closeBtn) {
        closeBtn.addEventListener('click', closeCameraTypeModal);
    }
});