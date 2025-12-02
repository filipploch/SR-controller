async function checkStatus() {
    try {
        const response = await fetch('/api/settings/status');
        const status = await response.json();
        
        // Aktualizuj status sezonu
        const seasonDot = document.getElementById('seasonStatus');
        if (status.has_current_season) {
            seasonDot.classList.add('status-ok');
        } else {
            if (seasonDot.classList.contains('status-ok')){
                seasonDot.classList.remove('status-ok');   
            }
        }

        // Aktualizuj status odcinka
        const episodeDot = document.getElementById('episodeStatus');
        if (status.has_current_episode) {
            episodeDot.classList.add('status-ok');
        } else {
            if (episodeDot.classList.contains('status-ok')){
                episodeDot.classList.remove('status-ok');   
            }
        }

        // Włącz/wyłącz link do kontrolera
        const controllerLink = document.getElementById('controllerLink');
        const warningContainer = document.getElementById('warningContainer');

        if (!status.can_access_controller) {
            controllerLink.classList.add('disabled');
            warningContainer.innerHTML = `
                <div class="warning-message">
                    <strong>⚠️ Kontroler niedostępny</strong>
                    <p>Aby uruchomić kontroler, musisz najpierw:</p>
                    <ul>
                        ${!status.has_current_season ? '<li>Utworzyć sezon i oznaczyć go jako aktualny w <a href="/seasons">Zarządzaniu Sezonami</a></li>' : ''}
                        ${!status.has_current_episode ? '<li>Utworzyć odcinek i oznaczyć go jako aktualny w <a href="/episodes">Zarządzaniu Odcinkami</a></li>' : ''}
                    </ul>
                </div>
            `;
        } else {
            controllerLink.classList.remove('disabled');
            warningContainer.innerHTML = '';
        }
    } catch (error) {
        console.error('Błąd pobierania statusu:', error);
    }
}

// Sprawdź status przy ładowaniu strony
checkStatus();

// Odświeżaj status co 5 sekund
setInterval(checkStatus, 5000);