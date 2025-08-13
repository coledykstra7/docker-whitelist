// JavaScript for Squid Proxy List Editor

function reloadSquid() {
    fetch('/reload', { method: 'POST' })
        .then(res => res.json())
        .then(data => {
            document.getElementById('squidStatus').textContent = data.status || 'reloaded';
        });
}

function clearAllLogs() {
    if (!confirm('Clear all access logs (WL, BL, and RG)? This cannot be undone.')) {
        return;
    }
    
    fetch('/clear-all-logs', { method: 'POST' })
        .then(res => res.json())
        .then(data => {
            alert(data.status || 'All logs cleared');
            updateSummary();
            updateLog();
        })
        .catch(err => {
            console.error('Error clearing logs:', err);
            alert('Error clearing logs: ' + err.message);
        });
}

function updateSummary() {
    fetch('/summary-data')
        .then(res => res.json())
        .then(data => {
            renderFilteredSummary(data.rows);
        });
}

function renderFilteredSummary(rows) {
    const showWL = document.getElementById('filterWL').checked;
    const showBL = document.getElementById('filterBL').checked;
    const showRG = document.getElementById('filterRG').checked;
    
    // Filter rows based on checkbox states
    const filteredRows = rows.filter(row => {
        if (row.status === '✅' && !showWL) return false;
        if (row.status === '❌' && !showBL) return false;
        if (row.status === '❓' && !showRG) return false;
        return true;
    });
    
    // Build HTML table
    let html = '<table class="summary-table"><tr><th>Actions</th><th>Domain</th><th>Count</th><th>Status</th></tr>';
    
    filteredRows.forEach(row => {
        let cls = "unknown";
        if (row.status === "✅") {
            cls = "whitelist";
        } else if (row.status === "❌") {
            cls = "blacklist";
        }
        
        // Generate action buttons based on current status
        let actions = "";
        const domain = escapeHtml(row.domain);
        if (row.status === "✅") {
            // Whitelisted: can move to blacklist
            actions = `<button onclick="moveDomain('${domain}', 'blacklist')" class="action-btn bl">→ BL</button>`;
        } else if (row.status === "❌") {
            // Blacklisted: can move to whitelist
            actions = `<button onclick="moveDomain('${domain}', 'whitelist')" class="action-btn wl">→ WL</button>`;
        } else {
            // Unknown: can move to whitelist or blacklist
            actions = `<button onclick="moveDomain('${domain}', 'whitelist')" class="action-btn wl">→ WL</button> <button onclick="moveDomain('${domain}', 'blacklist')" class="action-btn bl">→ BL</button>`;
        }
        
        html += `<tr><td>${actions}</td><td>${domain}</td><td>${row.count}</td><td class="status ${cls}">${row.status}</td></tr>`;
    });
    
    html += '</table>';
    document.getElementById('summary-content').innerHTML = html;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
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

function updateLists() {
    fetch('/lists')
        .then(res => res.json())
        .then(data => {
            // Update hidden textareas for form submission
            document.getElementById('whitelist-hidden').value = data.whitelist;
            document.getElementById('blacklist-hidden').value = data.blacklist;
            
            // Update table displays
            renderListTable('whitelist', data.whitelist);
            renderListTable('blacklist', data.blacklist);
        })
        .catch(err => {
            console.error('Error updating lists:', err);
        });
}

function renderListTable(listType, content) {
    const tableId = listType + '-table';
    const table = document.getElementById(tableId);
    
    // Clear existing rows except header
    while (table.rows.length > 1) {
        table.deleteRow(1);
    }
    
    // Parse content into entries
    const entries = parseListContent(content);
    
    // Add rows for each entry
    entries.forEach(entry => {
        const row = table.insertRow();
        
        // Create cells
        const actionsCell = row.insertCell(0);
        const domainCell = row.insertCell(1);
        const noteCell = row.insertCell(2);
        
        // Set cell classes
        actionsCell.className = 'actions-col';
        domainCell.className = 'domain-col';
        noteCell.className = 'note-col';
        
        // Set content
        domainCell.textContent = entry.domain;
        noteCell.textContent = entry.note;
        
        // Create buttons programmatically to avoid escaping issues
        const moveBtn = document.createElement('button');
        moveBtn.className = listType === 'whitelist' ? 'action-btn bl' : 'action-btn wl';
        moveBtn.textContent = listType === 'whitelist' ? '→ BL' : '→ WL';
        moveBtn.onclick = () => moveFromList(entry.domain, listType, listType === 'whitelist' ? 'blacklist' : 'whitelist');
        
        const removeBtn = document.createElement('button');
        removeBtn.className = 'remove-btn';
        removeBtn.textContent = '✕';
        removeBtn.onclick = () => removeFromList(entry.domain, listType);
        
        actionsCell.appendChild(moveBtn);
        actionsCell.appendChild(document.createTextNode(' '));
        actionsCell.appendChild(removeBtn);
    });
}

function parseListContent(content) {
    const entries = [];
    const lines = content.split('\n');
    
    lines.forEach(line => {
        line = line.trim();
        if (line && !line.startsWith('#')) {
            const parts = line.split('#');
            const domain = parts[0].trim();
            const note = parts.length > 1 ? parts.slice(1).join('#').trim() : '';
            if (domain) {
                entries.push({ domain, note });
            }
        }
    });
    
    return entries;
}

function addToList(listType) {
    const domainInput = document.getElementById(`new-${listType === 'whitelist' ? 'wl' : 'bl'}-domain`);
    const noteInput = document.getElementById(`new-${listType === 'whitelist' ? 'wl' : 'bl'}-note`);
    
    const domain = domainInput.value.trim();
    const note = noteInput.value.trim();
    
    if (!domain) {
        alert('Please enter a domain');
        return;
    }
    
    const data = new FormData();
    data.append('domain', domain);
    data.append('target', listType);
    data.append('note', note);
    
    fetch('/move-domain', { 
        method: 'POST',
        body: data
    })
    .then(res => res.json())
    .then(data => {
        if (data.status === 'success') {
            // Clear inputs
            domainInput.value = '';
            noteInput.value = '';
            
            // Refresh displays
            updateSummary();
            updateLog();
            updateLists();
        } else {
            alert('Error: ' + (data.error || 'Failed to add domain'));
        }
    })
    .catch(err => {
        console.error('Error adding domain:', err);
        alert('Error adding domain: ' + err.message);
    });
}

function moveFromList(domain, fromList, toList) {
    const data = new FormData();
    data.append('domain', domain);
    data.append('target', toList);
    data.append('note', ''); // No note when moving existing entries
    
    fetch('/move-domain', { 
        method: 'POST',
        body: data
    })
    .then(res => res.json())
    .then(data => {
        if (data.status === 'success') {
            updateSummary();
            updateLog();
            updateLists();
        } else {
            alert('Error: ' + (data.error || 'Failed to move domain'));
        }
    })
    .catch(err => {
        console.error('Error moving domain:', err);
        alert('Error moving domain: ' + err.message);
    });
}

function removeFromList(domain, fromList) {
    const data = new FormData();
    data.append('domain', domain);
    data.append('target', 'unknown'); // Remove from both lists
    data.append('note', '');
    
    fetch('/move-domain', { 
        method: 'POST',
        body: data
    })
    .then(res => res.json())
    .then(data => {
        if (data.status === 'success') {
            updateSummary();
            updateLog();
            updateLists();
        } else {
            alert('Error: ' + (data.error || 'Failed to remove domain'));
        }
    })
    .catch(err => {
        console.error('Error removing domain:', err);
        alert('Error removing domain: ' + err.message);
    });
}

function updateHiddenTextareas() {
    // Sync table data to hidden textareas before form submission
    const whitelistEntries = [];
    const blacklistEntries = [];
    
    // Extract whitelist data
    const wlTable = document.getElementById('whitelist-table');
    for (let i = 1; i < wlTable.rows.length; i++) {
        const domain = wlTable.rows[i].cells[1].textContent; // Domain is now column 1
        const note = wlTable.rows[i].cells[2].textContent;   // Note is now column 2
        if (note) {
            whitelistEntries.push(`${domain} #${note}`);
        } else {
            whitelistEntries.push(domain);
        }
    }
    
    // Extract blacklist data
    const blTable = document.getElementById('blacklist-table');
    for (let i = 1; i < blTable.rows.length; i++) {
        const domain = blTable.rows[i].cells[1].textContent; // Domain is now column 1
        const note = blTable.rows[i].cells[2].textContent;   // Note is now column 2
        if (note) {
            blacklistEntries.push(`${domain} #${note}`);
        } else {
            blacklistEntries.push(domain);
        }
    }
    
    document.getElementById('whitelist-hidden').value = whitelistEntries.join('\n');
    document.getElementById('blacklist-hidden').value = blacklistEntries.join('\n');
    
    return true; // Allow form submission
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

function setupFilterControls() {
    const filterWL = document.getElementById('filterWL');
    const filterBL = document.getElementById('filterBL');
    const filterRG = document.getElementById('filterRG');
    
    // Add event listeners to filter checkboxes
    [filterWL, filterBL, filterRG].forEach(checkbox => {
        checkbox.addEventListener('change', () => {
            // Re-fetch and re-render with current filter settings
            fetch('/summary-data')
                .then(res => res.json())
                .then(data => {
                    renderFilteredSummary(data.rows);
                });
        });
    });
}

function moveDomain(domain, targetStatus) {
    const noteField = document.getElementById('domainNote');
    const note = noteField.value.trim();
    const data = new FormData();
    data.append('domain', domain);
    data.append('target', targetStatus);
    data.append('note', note);
    
    fetch('/move-domain', { 
        method: 'POST',
        body: data
    })
    .then(res => res.json())
    .then(data => {
        if (data.status === 'success') {
            // Refresh the summary, log, and the whitelist/blacklist textareas
            updateSummary();
            updateLog();
            updateLists();
            
            // Optional: Clear note after successful move (uncomment if desired)
            // noteField.value = '';
            // localStorage.setItem('squidEditorDomainNote', '');
        } else {
            alert('Error: ' + (data.error || 'Failed to move domain'));
        }
    })
    .catch(err => {
        console.error('Error moving domain:', err);
        alert('Error moving domain: ' + err.message);
    });
}

document.addEventListener('DOMContentLoaded', function() {
    updateSummary();
    updateLog();
    updateLists();
    setupAutoRefresh();
    setupFilterControls();
    setupNotePersistence();
});

function setupNotePersistence() {
    const noteField = document.getElementById('domainNote');
    const storageKey = 'squidEditorDomainNote';
    
    // Load saved note on page load
    const savedNote = localStorage.getItem(storageKey);
    if (savedNote) {
        noteField.value = savedNote;
    }
    
    // Save note whenever it changes
    noteField.addEventListener('input', function() {
        localStorage.setItem(storageKey, noteField.value);
    });
    
    // Also save on blur to catch copy/paste operations
    noteField.addEventListener('blur', function() {
        localStorage.setItem(storageKey, noteField.value);
    });
}
