// JavaScript for Squid Proxy List Editor

function reloadSquid() {
    fetch('/reload', { method: 'POST' })
        .then(res => res.json())
        .then(data => {
            document.getElementById('squidStatus').textContent = data.status || 'reloaded';
        });
}

function setSetpoint() {
    fetch('/setpoint', { method: 'POST' })
        .then(res => res.json())
        .then(data => {
            document.getElementById('setpointDisplay').textContent = data.setpoint || 'set';
        });
}

function clearSetpoint() {
    fetch('/setpoint', { method: 'DELETE' })
        .then(res => res.json())
        .then(data => {
            document.getElementById('setpointDisplay').textContent = 'none';
        });
}

function updateSummary() {
    fetch('/summary')
        .then(res => res.text())
        .then(html => {
            document.getElementById('summary-box').innerHTML = html;
        });
}

function updateLog() {
    fetch('/log')
        .then(res => res.text())
        .then(text => {
            document.getElementById('log').textContent = text;
        });
}

function setupAutoRefresh() {
    let autoRefresh = document.getElementById('autoRefresh');
    let intervalId = null;
    function refresh() {
        updateSummary();
        updateLog();
    }
    autoRefresh.addEventListener('change', function() {
        if (autoRefresh.checked) {
            intervalId = setInterval(refresh, 5000);
        } else {
            clearInterval(intervalId);
        }
    });
    if (autoRefresh.checked) {
        intervalId = setInterval(refresh, 5000);
    }
}

document.addEventListener('DOMContentLoaded', function() {
    updateSummary();
    updateLog();
    setupAutoRefresh();
});
