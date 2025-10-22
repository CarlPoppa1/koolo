const supervisorNameElement = document.getElementById('supervisor-name');
const logContainer = document.getElementById('log-container');
const autoScrollCheckbox = document.getElementById('auto-scroll');
const filterButtons = document.querySelectorAll('.filter-btn');
const searchInput = document.getElementById('search-input');
const searchPrevBtn = document.getElementById('search-prev-btn');
const searchNextBtn = document.getElementById('search-next-btn');
const searchResults = document.getElementById('search-results');
const maxLinesInput = document.getElementById('max-lines');
const copyDataBtn = document.getElementById('copy-data-btn');

let refreshInterval = 1000; // 1 second like debug screen
let refreshIntervalId = null;
let currentOffset = 0;
let isPageVisible = true;
let currentFilter = 'debug';
let maxLogLines = parseInt(localStorage.getItem('maxLogLines') || '1000'); // Default to 1000 lines, 0 = unlimited

const levelHierarchy = ['trace', 'debug', 'info', 'warn', 'error'];

let searchMatches = [];
let currentSearchIndex = -1;
let lastSearchTerm = '';

const urlParams = new URLSearchParams(window.location.search);
const characterName = urlParams.get('characterName') || 'unknown';

supervisorNameElement.querySelector('span').textContent = characterName;

function fetchLogData() {
    if (!isPageVisible) {
        return;
    }

    fetch(`/logs-data?characterName=${encodeURIComponent(characterName)}&offset=${currentOffset}`)
        .then(response => response.json())
        .then(data => {
            if (data.offset !== undefined) {
                currentOffset = data.offset;
            }

            if (data.content && data.content.length > 0) {
                if (data.isInitial) {
                    logContainer.innerHTML = '';
                    appendLogContent(data.content);
                } else {
                    appendLogContent(data.content);
                }

                if (lastSearchTerm) {
                    performSearch(lastSearchTerm, false);
                }
            }
        })
        .catch(error => {
            console.error('Error fetching log data:', error);
            logContainer.innerHTML = '<div style="color: #ff6b6b;">Error loading logs. Please refresh the page.</div>';
        });
}

function appendLogContent(content) {
    const lines = content.split('\n');
    lines.forEach(line => {
        if (line.trim()) {
            appendLogLine(line);
        }
    });
}

function appendLogLine(line) {
    const logLine = document.createElement('div');
    logLine.className = 'log-line';

    let lineLevel = null;

    try {
        const logObj = JSON.parse(line);
        if (logObj.level) {
            const level = logObj.level.toLowerCase();
            logLine.classList.add(level);
            lineLevel = level;
        }
        const time = logObj.time || '';
        const level = logObj.level || '';
        const msg = logObj.msg || '';
        logLine.textContent = `[${time}] ${level}: ${msg}`;
    } catch (e) {
        logLine.textContent = line;
    }

    if (lineLevel && !shouldShowLevel(lineLevel)) {
        logLine.classList.add('hidden');
    }

    logContainer.appendChild(logLine);

    if (maxLogLines > 0) {
        const allLines = logContainer.querySelectorAll('.log-line');
        if (allLines.length > maxLogLines) {
            const linesToRemove = allLines.length - maxLogLines;
            for (let i = 0; i < linesToRemove; i++) {
                logContainer.removeChild(allLines[i]);
            }
        }
    }

    if (autoScrollCheckbox.checked) {
        logContainer.scrollTop = logContainer.scrollHeight;
    }
}

function shouldShowLevel(level) {
    const filterIndex = levelHierarchy.indexOf(currentFilter);
    const levelIndex = levelHierarchy.indexOf(level);

    return levelIndex >= filterIndex;
}

function applyFilter() {
    const allLogLines = logContainer.querySelectorAll('.log-line');

    allLogLines.forEach(logLine => {
        let lineLevel = null;
        for (const level of levelHierarchy) {
            if (logLine.classList.contains(level)) {
                lineLevel = level;
                break;
            }
        }

        if (lineLevel) {
            if (shouldShowLevel(lineLevel)) {
                logLine.classList.remove('hidden');
            } else {
                logLine.classList.add('hidden');
            }
        }
    });

    // This prevents auto-scroll from being unchecked when filter changes cause height changes
    if (autoScrollCheckbox.checked) {
        logContainer.scrollTop = logContainer.scrollHeight;
    }
}

function updateFilterButtonStates() {
    const filterIndex = levelHierarchy.indexOf(currentFilter);
    filterButtons.forEach(btn => {
        const btnLevel = btn.dataset.level;
        const btnIndex = levelHierarchy.indexOf(btnLevel);

        if (btnIndex >= filterIndex) {
            btn.classList.add('active');
        } else {
            btn.classList.remove('active');
        }
    });
}

function setupFilterButtons() {
    filterButtons.forEach(button => {
        button.addEventListener('click', function() {
            currentFilter = this.dataset.level;
            updateFilterButtonStates();
            applyFilter();
        });
    });

    updateFilterButtonStates();
}

function performSearch(searchTerm, scrollToFirst) {
    lastSearchTerm = searchTerm;

    const allLogLines = logContainer.querySelectorAll('.log-line');
    allLogLines.forEach(line => {
        line.classList.remove('highlight', 'current-highlight');
    });

    searchMatches = [];

    if (!searchTerm) {
        searchResults.textContent = '0/0';
        currentSearchIndex = -1;
        return;
    }

    const searchLower = searchTerm.toLowerCase();
    allLogLines.forEach((line, index) => {
        if (!line.classList.contains('hidden') && line.textContent.toLowerCase().includes(searchLower)) {
            line.classList.add('highlight');
            searchMatches.push(line);
        }
    });

    if (searchMatches.length > 0) {
        currentSearchIndex = 0;
        searchMatches[0].classList.add('current-highlight');
        searchResults.textContent = `1/${searchMatches.length}`;

        if (scrollToFirst) {
            searchMatches[0].scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
    } else {
        searchResults.textContent = '0/0';
        currentSearchIndex = -1;
    }
}

function navigateSearch(direction) {
    if (searchMatches.length === 0) return;

    if (currentSearchIndex >= 0) {
        searchMatches[currentSearchIndex].classList.remove('current-highlight');
    }

    if (direction === 'next') {
        currentSearchIndex = (currentSearchIndex + 1) % searchMatches.length;
    } else {
        currentSearchIndex = (currentSearchIndex - 1 + searchMatches.length) % searchMatches.length;
    }

    searchMatches[currentSearchIndex].classList.add('current-highlight');
    searchMatches[currentSearchIndex].scrollIntoView({ behavior: 'smooth', block: 'center' });

    searchResults.textContent = `${currentSearchIndex + 1}/${searchMatches.length}`;
}

function setupSearch() {
    searchInput.addEventListener('input', function() {
        performSearch(this.value, true);
    });

    searchPrevBtn.addEventListener('click', function() {
        navigateSearch('prev');
    });

    searchNextBtn.addEventListener('click', function() {
        navigateSearch('next');
    });

    searchInput.addEventListener('keydown', function(e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            navigateSearch('next');
        }
    });
}

function handleVisibilityChange() {
    if (document.hidden) {
        isPageVisible = false;
        console.log('Page hidden - pausing log polling');
    } else {
        isPageVisible = true;
        console.log('Page visible - resuming log polling');
        fetchLogData();
    }
}

function handleMaxLinesChange() {
    const newValue = parseInt(maxLinesInput.value) || 0;
    maxLogLines = newValue;

    localStorage.setItem('maxLogLines', maxLogLines.toString());

    // Reset and refetch logs to apply the new limit properly
    // This ensures that when changing to unlimited (0), all logs are fetched
    currentOffset = 0;
    logContainer.innerHTML = '<div style="color: #888;">Reloading logs...</div>';

    lastSearchTerm = '';
    searchMatches = [];
    currentSearchIndex = -1;
    searchResults.textContent = '0/0';
    searchInput.value = '';

    fetchLogData();
}

function isScrolledToBottom() {
    const threshold = 50;
    return logContainer.scrollHeight - logContainer.scrollTop - logContainer.clientHeight < threshold;
}


function handleScroll() {
    if (!isScrolledToBottom() && autoScrollCheckbox.checked) {
        autoScrollCheckbox.checked = false;
    } else if (isScrolledToBottom() && !autoScrollCheckbox.checked) {

        autoScrollCheckbox.checked = true;
    }
}

function copyLogData() {
    const visibleLogLines = logContainer.querySelectorAll('.log-line:not(.hidden)');
    const logText = Array.from(visibleLogLines)
        .map(line => line.textContent)
        .join('\n');

    navigator.clipboard.writeText(logText).then(() => {
        const originalHTML = copyDataBtn.innerHTML;
        copyDataBtn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>Copied!';
        setTimeout(() => {
            copyDataBtn.innerHTML = originalHTML;
        }, 2000);
    }).catch(err => {
        console.error('Failed to copy log data:', err);
    });
}

window.onload = function () {
  setupFilterButtons();
  setupSearch();

  if (maxLinesInput) {
    maxLinesInput.value = String(maxLogLines);
    maxLinesInput.addEventListener('input', handleMaxLinesChange);
  }

  if (copyDataBtn) {
    copyDataBtn.addEventListener('click', copyLogData);
  }

  logContainer.addEventListener('scroll', handleScroll, { passive: true });
  document.addEventListener('visibilitychange', handleVisibilityChange);
  window.addEventListener('beforeunload', () => {
    if (refreshIntervalId) clearInterval(refreshIntervalId);
  });

  fetchLogData();
  refreshIntervalId = setInterval(fetchLogData, refreshInterval);
};