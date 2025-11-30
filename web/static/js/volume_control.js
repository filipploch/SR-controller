// volume_control.js - Kontrola głośności źródeł audio

// Nasłuchuj na zmiany głośności z OBS (zewnętrzne zmiany)
socket.on('volume_changed', (data) => {
    console.log('Volume changed in OBS:', data);
    
    // Znajdź suwak dla tego źródła
    const slider = document.querySelector(
        `.volume-slider[data-source-name="${data.source_name}"]`
    );
    
    if (slider) {
        // Zaktualizuj suwak bez triggerowania eventu
        slider.value = data.volume_db;
        
        // Oznacz wyciszone
        if (data.volume_db <= -100) {
            slider.classList.add('muted');
        } else {
            slider.classList.remove('muted');
        }
        
        console.log(`Updated slider for ${data.source_name} to ${data.volume_db}dB`);
    }
});

// Ustaw głośność źródła (wywołane przez suwak)
function setSourceVolume(sourceName, volumeDb) {
    socket.emit('set_input_volume', JSON.stringify({
        inputName: sourceName,
        inputVolumeDb: volumeDb
    }), (response) => {
        const data = JSON.parse(response);
        if (!data.success) {
            console.error('Failed to set volume:', data.error);
            alert('Nie udało się ustawić głośności: ' + data.error);
        }
    });
}

// Pobierz aktualną głośność z OBS
async function getSourceVolume(sourceName) {
    return new Promise((resolve) => {
        socket.emit('get_input_volume', sourceName, (response) => {
            const data = JSON.parse(response);
            console.log("data:", data);
            if (data.success && data) {
                resolve(data);
            } else {
                console.error('Failed to get volume:', data.error || 'unknown error');
                resolve(-10); // Domyślnie -10dB
            }
        });
    });
}

// Formatuj label głośności
function formatVolumeLabel(volumeDb) {
    if (volumeDb <= -100) {
        return '-∞';
    }
    return `${volumeDb.toFixed(1)}dB`;
}

// Renderuj źródło z suwakiem głośności
function renderSourceWithVolume(source, sceneName, container) {
    const sourceName = source.sourceName || source.source_name || 'Źródło';
    
    // Wrapper główny
    const mainWrapper = document.createElement('div');
    mainWrapper.className = 'source-with-volume';
    
    // Wrapper dla przycisku + przycisku modalu (jeśli mikrofon)
    const buttonWrapper = document.createElement('div');
    buttonWrapper.className = 'source-button-wrapper';
    
    // Główny przycisk źródła
    const button = document.createElement('button');
    button.className = 'source-btn with-volume';
    button.textContent = sourceName;
    button.dataset.sceneName = sceneName;
    button.dataset.sourceName = sourceName;
    button.dataset.sceneItemId = source.sceneItemId || 0;
    
    const isVisible = source.sceneItemEnabled !== undefined 
        ? source.sceneItemEnabled 
        : (source.is_visible || false);
        
    if (isVisible) {
        button.classList.add('active');
    }
    
    button.addEventListener('dblclick', () => {
        const isCurrentlyActive = button.classList.contains('active');
        
        if (MAIN_SCENES.includes(sceneName)) {
            if (isCurrentlyActive) return;
            switchMainSource(sceneName, button.dataset.sourceName, button.dataset.sceneItemId);
        } else {
            toggleSource(sceneName, button.dataset.sourceName, !isCurrentlyActive);
        }
    });
    
    buttonWrapper.appendChild(button);
    
    // Przycisk ▼ dla mikrofonów (przypisanie osoby)
    if (sceneName === 'MIKROFONY') {
        const modalButton = document.createElement('button');
        modalButton.className = 'open-modal-btn';
        modalButton.textContent = '▼';
        modalButton.title = 'Przypisz osobę';
        modalButton.onclick = (e) => {
            e.stopPropagation();
            openMicrophoneAssignModal(sourceName, sceneName);
        };
        buttonWrapper.appendChild(modalButton);
    }
    
    mainWrapper.appendChild(buttonWrapper);
    
    // Suwak głośności
    const sliderContainer = document.createElement('div');
    sliderContainer.className = 'volume-slider-container';
    
    const slider = document.createElement('input');
    slider.type = 'range';
    slider.className = 'volume-slider';
    slider.min = -100;  // -100 dB (prawie cisza)
    slider.max = 0;     // 0 dB (max)
    slider.value = -10; // Domyślnie -10dB
    slider.dataset.sourceName = sourceName;
    
    // Pobierz aktualną głośność z OBS
    try {
        const currentVol = getSourceVolume(sourceName);
        var currentVolume = currentVol.data['volume_db']
        slider.value = currentVolume;

        console.log("currentVolume: ", currentVolume);
        
        // Oznacz wyciszone
        if (currentVolume <= -100) {
            slider.classList.add('muted');
        }
    } catch (error) {
        console.error('Error getting initial volume:', error);
    }
    
    // Event - zmiana głośności
    slider.addEventListener('input', (e) => {
        const volumeDb = parseFloat(e.target.value);
        
        // Oznacz wyciszone
        if (volumeDb <= -100) {
            slider.classList.add('muted');
        } else {
            slider.classList.remove('muted');
        }
    });
    
    // Event - po zakończeniu zmiany (mouseup/touchend)
    slider.addEventListener('change', (e) => {
        const volumeDb = parseFloat(e.target.value);
        setSourceVolume(sourceName, volumeDb);
    });
    
    sliderContainer.appendChild(slider);
    
    mainWrapper.appendChild(sliderContainer);
    
    return mainWrapper;
}

// Pomocnicza: sprawdź czy źródło powinno mieć suwak
function shouldHaveVolumeSlider(sourceName, sceneName) {
    // MIKROFONY - wszystkie źródła
    if (sceneName === 'MIKROFONY') {
        return true;
    }
    
    // MUZYKA - wszystkie źródła
    if (sceneName === 'MUZYKA') {
        return true;
    }
    
    return false;
}

// Placeholder: Otwórz modal przypisania osoby do mikrofonu
function openMicrophoneAssignModal(sourceName, sceneName) {
    alert(`Modal przypisania osoby do mikrofonu: ${sourceName}\n(Do zaimplementowania w następnym kroku)`);
    // TODO: Implementacja modalu przypisania Staff/Guest do mikrofonu
}