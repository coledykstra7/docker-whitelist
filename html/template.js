// JavaScript for Squid Proxy List Editor

function reloadSquid() {
    fetch('/reload', { method: 'POST' })
        .then(res => res.json())
        .then(data => {
            document.getElementById('squidStatus').textContent = data.status || 'reloaded';
        });
}

function clearWhitelistLog() {
    fetch('/clear-whitelist', { method: 'POST' })
        .then(res => res.json())
        .then(data => {
            alert(data.status || 'Whitelist log cleared');
            updateSummary();
            updateLog();
        })
        .catch(err => {
            console.error('Error clearing whitelist log:', err);
        });
}

function clearBlacklistLog() {
    fetch('/clear-blacklist', { method: 'POST' })
        .then(res => res.json())
        .then(data => {
            alert(data.status || 'Blacklist log cleared');
            updateSummary();
            updateLog();
        })
        .catch(err => {
            console.error('Error clearing blacklist log:', err);
        });
}

function clearRegularLog() {
    fetch('/clear-regular', { method: 'POST' })
        .then(res => res.json())
        .then(data => {
            alert(data.status || 'Regular log cleared');
            updateSummary();
            updateLog();
        })
        .catch(err => {
            console.error('Error clearing regular log:', err);
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
            const logElement = document.getElementById('log');
            // Split text into lines and process each line for color coding
            const lines = text.split('\n');
            const processedLines = lines.map(line => {
                if (line.trim() === '') return line;
                
                // Check for WL, BL, or RG tags (should be second field after timestamp)
                const fields = line.split(' ');
                if (fields.length >= 2) {
                    const tag = fields[1];
                    if (tag === 'WL') {
                        return `<span class="log-WL">${line}</span>`;
                    } else if (tag === 'BL') {
                        return `<span class="log-BL">${line}</span>`;
                    } else if (tag === 'RG') {
                        return `<span class="log-RG">${line}</span>`;
                    }
                }
                return line;
            });
            
            logElement.innerHTML = processedLines.join('\n');
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
