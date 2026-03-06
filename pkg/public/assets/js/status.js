const history = [];
const MAX = 60;
async function fetchHealth() {
    try {
        const res = await fetch('/health');
        const data = await res.json();
        history.push(data);
        if (history.length > MAX) history.shift();
        render();
    } catch (e) {
        render(true);
    }
}

function render(error = false){
    const indicator = document.getElementById('indicator');
    const fields = document.getElementById('fields');

    const latest = history[history.length - 1];

    if (error || !latest) {
        indicator.className = 'down';
        fields.textContent ='unreachable';
        return;
    }

    indicator.className = latest.status === 'ok' ? 'up' : 'down';

    fields.innerHTML = `
    <span>status: ${latest.status}</span>
    <span>version: ${latest.version}</span>
    <span>uptime: ${latest.uptime_seconds}s</span>
    <span>rooms: ${latest.rooms}</span>
    <span>federation: ${latest.federation}</span>`;

    drawChart();
}

function drawChart() {
    const svg = document.getElementById('chart');
    const width = 300;
    const height = 80;

    if (history.length < 2) return;

    const values = history.map(e => e.uptime_seconds);
    const min = Math.min(...values);
    const max = Math.max(...values);

    const points = values.map((v, i) => {
        const x = (i / (values.length - 1)) * width;
        const y = max === min ? height / 2 : height - ((v - min) / (max - min)) * height;
        return `${x},${y}`;
    }).join(' ');

    svg.innerHTML = `
    <polyline
    points="${points}"
    fill="none"
    stroke="#4caf50"
    stroke-width="2"
    />`;
}

fetchHealth();
setInterval(fetchHealth, 30000);