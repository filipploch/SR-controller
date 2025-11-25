// Konfiguracja scen
const SCENES = ['KAMERY', 'MEDIA', 'REPORTAZE', 'MIKROFONY', 'MUZYKA'];
const MAIN_SCENES = ['KAMERY', 'MEDIA', 'REPORTAZE'];
const SCREEN_SCENE = 'SCREEN';
const SWITCH_DELAY = 600;

let currentActiveScene = null;
const socket = io();

const socketStatus = document.getElementById('socketStatus');
const obsStatus = document.getElementById('obsStatus');

// Funkcje przełączania zakładek
function switchTab(group, tabName) {
    const tabButton = document.querySelector(`.tab-btn[data-tab="${tabName}"]`);
    const tabContent = document.getElementById(`tab-${tabName}`);
    
    if (!tabButton || !tabContent) return;
    
    // Znajdź grupę zakładek
    const tabButtons = tabButton.parentElement;
    const tabsContainer = tabButtons.parentElement;
    
    // Wyłącz wszystkie przyciski i zawartość w tej grupie
    tabButtons.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    tabsContainer.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    
    // Włącz wybraną zakładkę
    tabButton.classList.add('active');
    tabContent.classList.add('active');
}

// Funkcje sterowania OBS
function obsStartStreaming() {
    socket.emit('obs_start_streaming', JSON.stringify({}), (response) => {
        console.log('Start streaming:', response);
    });
}

function obsStopStreaming() {
    socket.emit('obs_stop_streaming', JSON.stringify({}), (response) => {
        console.log('Stop streaming:', response);
    });
}

function obsStartRecording() {
    socket.emit('obs_start_recording', JSON.stringify({}), (response) => {
        console.log('Start recording:', response);
    });
}

function obsStopRecording() {
    socket.emit('obs_stop_recording', JSON.stringify({}), (response) => {
        console.log('Stop recording:', response);
    });
}

function obsToggleStudioMode() {
    socket.emit('obs_toggle_studio_mode', JSON.stringify({}), (response) => {
        console.log('Toggle studio mode:', response);
    });
}

function obsTransition() {
    socket.emit('obs_trigger_transition', JSON.stringify({}), (response) => {
        console.log('Trigger transition:', response);
    });
}

socket.on('connect', () => {
	console.log('Połączono z Socket.IO');
	socketStatus.classList.add('connected');
	loadAllScenes();
});

socket.on('disconnect', () => {
	console.log('Rozłączono z Socket.IO');
	socketStatus.classList.remove('connected');
});

socket.on('source_changed', (data) => {
	console.log('Zmieniono źródło:', data);
	updateSourceButton(data.scene_name, data.source_name, data.visible);
});

// Ładowanie aktualnego media dla przycisków Media1 i Reportaze1
async function loadCurrentMediaButtons() {
	await loadCurrentMediaButton('MEDIA', 'Media1');
	await loadCurrentMediaButton('REPORTAZE', 'Reportaze1');
}

async function loadCurrentMediaButton(sceneName, sourceName) {
	try {
		const response = await fetch(`/api/episodes/current/media/scene/${sceneName}`);
		const data = await response.json();
		
		const containerId = `sources-${sceneName.toLowerCase()}`;
		const container = document.getElementById(containerId);
		if (!container) return;
		
		// Znajdź przycisk dla tego źródła
		const button = container.querySelector(`[data-source-name="${sourceName}"]`);
		if (!button) return;
		
		if (data.success && data.title) {
			// Zaktualizuj tekst przycisku na tytuł z bazy
			button.textContent = data.title;
			
			// Zapisz ID media w atrybucie data
			button.dataset.mediaId = data.media_id;
		} else {
			// Brak media - przywróć oryginalną nazwę
			button.textContent = sourceName;
			button.removeAttribute('data-media-id');
		}
	} catch (error) {
		console.error(`Błąd ładowania media dla ${sceneName}:`, error);
	}
}

function loadAllScenes() {
	SCENES.forEach(sceneName => {
		loadSceneSources(sceneName);
	});
	detectActiveScene();
}

function detectActiveScene() {
	MAIN_SCENES.forEach(sceneName => {
		const containerId = `sources-${sceneName.toLowerCase()}`;
		const container = document.getElementById(containerId);
		if (container) {
			const activeButton = container.querySelector('.source-btn.active');
			if (activeButton) {
				currentActiveScene = sceneName;
			}
		}
	});
}

function sendToOverlay(action, params = {}) {
	socket.emit('send_to_overlay', JSON.stringify({
		action: action,
		...params
	}));
}

function loadSceneSources(sceneName) {
	// Najpierw synchronizuj kolejność z bazy do OBS
	socket.emit('sync_source_order', JSON.stringify({
		scene_name: sceneName
	}), (syncResponse) => {
		// Po synchronizacji pobierz źródła
		socket.emit('get_sources', sceneName, (response) => {
			try {
				const data = JSON.parse(response);
				if (data.success) {
					renderSources(sceneName, data.data.sources);
					if (data.data.has_changes) {
						showSaveButton(sceneName);
					}
					obsStatus.classList.add('connected');
				}
			} catch (error) {
				console.error('Błąd:', error);
			}
		});
	});
}

function showSaveButton(sceneName) {
	const buttonId = `save-${sceneName.toLowerCase()}`;
	const button = document.getElementById(buttonId);
	if (button) {
		button.classList.add('visible');
	}
}

function hideSaveButton(sceneName) {
	const buttonId = `save-${sceneName.toLowerCase()}`;
	const button = document.getElementById(buttonId);
	if (button) {
		button.classList.remove('visible');
	}
}

function saveSourceOrder(sceneName) {
	socket.emit('save_source_order', JSON.stringify({
		scene_name: sceneName
	}), (response) => {
		const data = JSON.parse(response);
		alert(`Zapisano kolejność dla sceny ${sceneName}`);
		hideSaveButton(sceneName);
	});
}

function renderSources(sceneName, sources) {
	const containerId = `sources-${sceneName.toLowerCase()}`;
	const container = document.getElementById(containerId);
	
	if (!container) return;
	container.innerHTML = '';
	
	if (!sources || sources.length === 0) {
		container.innerHTML = '<div class="loading">Brak źródeł</div>';
		return;
	}
	
	const reversedSources = [...sources].reverse();
	
	reversedSources.forEach(source => {
		const button = document.createElement('button');
		button.className = 'source-btn';
		
		const sourceName = source.sourceName || source.source_name || 'Źródło';
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
		
		container.appendChild(button);
	});
	
	// Po wyrenderowaniu źródeł, załaduj aktualne media dla Media1 i Reportaze1
	if (sceneName === 'MEDIA' || sceneName === 'REPORTAZE') {
		const mediaSourceName = sceneName === 'MEDIA' ? 'Media1' : 'Reportaze1';
		loadCurrentMediaButton(sceneName, mediaSourceName);
	}
}

function switchMainSource(sceneName, sourceName, sceneItemId) {
	const shouldShowTransition = !(currentActiveScene === 'KAMERY' && sceneName === 'KAMERY');
	
	if (shouldShowTransition) {
		sendToOverlay('show_transition');
	}
	
	currentActiveScene = sceneName;
	
	socket.emit('set_current_scene', JSON.stringify({
		scene_name: 'STREAM'
	}), () => {
		socket.emit('toggle_source', JSON.stringify({
			scene_name: sceneName,
			source_name: sourceName,
			visible: true
		}), (response) => {
			const data = JSON.parse(response);
			if (!data.success) {
				alert('Błąd: ' + data.error);
				return;
			}
			
			setTimeout(() => {
				socket.emit('set_source_index', JSON.stringify({
					scene_name: sceneName,
					source_name: sourceName,
					to_top: true
				}), () => {
					socket.emit('set_source_index', JSON.stringify({
						scene_name: SCREEN_SCENE,
						source_name: sceneName,
						to_top: true
					}), () => {
						turnOffAllMainScenes(sceneName, sourceName);
						manageMicrophones(sceneName);
						updateSourceButton(sceneName, sourceName, true);
					});
				});
			}, SWITCH_DELAY);
		});
	});
}

// Zarządzanie mikrofonami w zależności od sceny
function manageMicrophones(sceneName) {
	if (sceneName === 'REPORTAZE') {
		// Wyłącz wszystkie mikrofony przy reportażu
		console.log('Wyłączam wszystkie mikrofony (reportaż)');
		socket.emit('mute_all_microphones', JSON.stringify({}), (response) => {
			console.log('Mikrofony wyłączone:', response);
		});
	} else if (sceneName === 'KAMERY') {
		// Przywróć mikrofony z is_visible = true przy kamerach
		console.log('Przywracam aktywne mikrofony (kamery)');
		socket.emit('restore_microphones', JSON.stringify({}), (response) => {
			console.log('Mikrofony przywrócone:', response);
		});
	}
}

function turnOffAllMainScenes(exceptScene, exceptSource) {
	MAIN_SCENES.forEach(sceneName => {
		const containerId = `sources-${sceneName.toLowerCase()}`;
		const container = document.getElementById(containerId);
		if (!container) return;
		
		const buttons = container.querySelectorAll('.source-btn.active');
		buttons.forEach(button => {
			if (button.dataset.sceneName === exceptScene && 
				button.dataset.sourceName === exceptSource) {
				return;
			}
			toggleSource(button.dataset.sceneName, button.dataset.sourceName, false);
		});
	});
}

function toggleSource(sceneName, sourceName, visible) {
	socket.emit('toggle_source', JSON.stringify({
		scene_name: sceneName,
		source_name: sourceName,
		visible: visible
	}), (response) => {
		const data = JSON.parse(response);
		if (data.success) {
			updateSourceButton(sceneName, sourceName, visible);
		}
	});
}

function updateSourceButton(sceneName, sourceName, visible) {
	const containerId = `sources-${sceneName.toLowerCase()}`;
	const container = document.getElementById(containerId);
	if (!container) return;
	
	const buttons = container.querySelectorAll('.source-btn');
	buttons.forEach(button => {
		if (button.dataset.sourceName === sourceName) {
			if (visible) {
				button.classList.add('active');
			} else {
				button.classList.remove('active');
			}
		}
	});
}