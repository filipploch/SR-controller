const socket = io();
        
socket.on('connect', () => {
    console.log('Overlay połączony');
});

socket.on('overlay_message', (data) => {
    console.log('Wiadomość:', data);

    // data może zawierać: { action: 'nazwa_akcji', params: {...} }
    switch(data.action) {
        case 'show_transition':
            showTransition();
            break;
            
        default:
            console.log('Nieznana akcja:', data.action);
    }
});

// Funkcja do wyświetlenia przejścia
function showTransition() {
    const transitionBox = document.getElementById('transitionBox');
    if (transitionBox) {
        transitionBox.style.display = 'flex';
        
        setTimeout(() => {
            transitionBox.style.display = 'none';
        }, 2000);
    }
}