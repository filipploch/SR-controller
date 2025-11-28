// cameras.js - Zarządzanie typami kamer

let cameraTypes = [];
let editingId = null;

// Załaduj typy kamer
async function loadCameraTypes() {
    try {
        const response = await fetch('/api/camera-types');
        cameraTypes = await response.json();
        renderCameraTypes();
    } catch (error) {
        console.error('Błąd ładowania typów kamer:', error);
        document.getElementById('cameraTypesContainer').innerHTML = 
            '<div class="error">Błąd ładowania danych</div>';
    }
}

// Renderuj listę typów
function renderCameraTypes() {
    const container = document.getElementById('cameraTypesContainer');
    
    if (cameraTypes.length === 0) {
        container.innerHTML = '<div class="empty">Brak typów kamer</div>';
        return;
    }

    container.innerHTML = cameraTypes.map(type => `
        <div class="list-item ${type.is_system ? 'system' : ''}">
            <div class="list-item-info">
                <span class="list-item-order">#${type.order}</span>
                <span class="list-item-name">${type.name}</span>
                ${type.is_system ? '<span class="list-item-badge">Systemowy</span>' : ''}
            </div>
            <div class="list-item-actions">
                ${!type.is_system ? `
                    <button class="btn btn-small" onclick="editCameraType(${type.id})">
                        ✏️
                    </button>
                    <button class="btn btn-small btn-danger" onclick="deleteCameraType(${type.id}, '${type.name}')">
                        🗑️
                    </button>
                ` : `
                    <span style="color: #4CAF50; font-size: 9px;">🔒 Chroniony</span>
                `}
            </div>
        </div>
    `).join('');
}

// Otwórz modal dodawania
function openCameraTypeModal() {
    editingId = null;
    document.getElementById('modalTitle').textContent = 'Nowy Typ Kamery';
    document.getElementById('cameraTypeForm').reset();
    document.getElementById('cameraTypeId').value = '';
    document.getElementById('cameraTypeModal').classList.add('active');
}

// Zamknij modal
function closeCameraTypeModal() {
    document.getElementById('cameraTypeModal').classList.remove('active');
    editingId = null;
}

// Edytuj typ
function editCameraType(id) {
    const type = cameraTypes.find(t => t.id === id);
    if (!type) return;

    if (type.is_system) {
        alert('Nie można edytować systemowego typu kamery');
        return;
    }

    editingId = id;
    document.getElementById('modalTitle').textContent = 'Edytuj Typ Kamery';
    document.getElementById('cameraTypeId').value = type.id;
    document.getElementById('cameraTypeName').value = type.name;
    document.getElementById('cameraTypeModal').classList.add('active');
}

// Zapisz typ
async function saveCameraType(event) {
    event.preventDefault();

    const name = document.getElementById('cameraTypeName').value.trim();
    
    if (!name) {
        alert('Nazwa typu jest wymagana');
        return;
    }

    const data = { name };

    try {
        let response;
        
        if (editingId) {
            // Edycja
            response = await fetch(`/api/camera-types/${editingId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
        } else {
            // Nowy
            response = await fetch('/api/camera-types', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
        }

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        closeCameraTypeModal();
        loadCameraTypes();
        
        const action = editingId ? 'zaktualizowano' : 'dodano';
        console.log(`Typ kamery ${action} pomyślnie`);
        
    } catch (error) {
        console.error('Błąd zapisywania typu:', error);
        alert('Błąd: ' + error.message);
    }
}

// Usuń typ
async function deleteCameraType(id, name) {
    const type = cameraTypes.find(t => t.id === id);
    
    if (type && type.is_system) {
        alert('Nie można usunąć systemowego typu kamery');
        return;
    }

    if (!confirm(`Czy na pewno chcesz usunąć typ "${name}"?`)) {
        return;
    }

    try {
        const response = await fetch(`/api/camera-types/${id}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        loadCameraTypes();
        console.log('Typ kamery usunięty pomyślnie');
        
    } catch (error) {
        console.error('Błąd usuwania typu:', error);
        
        if (error.message.includes('assigned cameras')) {
            alert('Nie można usunąć typu, ponieważ są do niego przypisane kamery');
        } else {
            alert('Błąd usuwania typu: ' + error.message);
        }
    }
}

// Inicjalizacja
document.addEventListener('DOMContentLoaded', () => {
    loadCameraTypes();
});
