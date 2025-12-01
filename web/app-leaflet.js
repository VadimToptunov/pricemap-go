// Configuration
const API_BASE_URL = 'http://localhost:3000/api/v1';
let map;
let heatmapLayer;
let heatmapData = [];
let markers = [];
let propertiesData = [];
let factorsChart = null;
let autoUpdateInterval = null;
let debounceTimer = null;
let currentSelectedPoint = null;

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    initApp();
});

// Initialize application
function initApp() {
    initMap();
    initEventListeners();
    loadInitialData();
    startAutoUpdate();
}

// Initialize map (Leaflet instead of Google Maps)
function initMap() {
    map = L.map('map').setView([55.7558, 37.6173], 10); // Moscow by default
    
    // Add OpenStreetMap tiles (free, no API key required)
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors',
        maxZoom: 19
    }).addTo(map);
    
    // Load data when map bounds change
    map.on('moveend', () => {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(() => {
            loadHeatmapData();
        }, 500);
    });
    
    // Load data when zoom changes
    map.on('zoomend', () => {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(() => {
            loadHeatmapData();
        }, 300);
    });
}

// Initialize event listeners
function initEventListeners() {
    // Filters
    document.getElementById('applyFilters').addEventListener('click', applyFilters);
    document.getElementById('resetFilters').addEventListener('click', resetFilters);
    
    // Auto-update
    document.getElementById('autoUpdate').addEventListener('click', toggleAutoUpdate);
    
    // Details
    document.getElementById('toggleDetails').addEventListener('click', toggleDetailsPanel);
    document.getElementById('closePanel').addEventListener('click', closeInfoPanel);
    
    // Map controls
    document.getElementById('zoomIn').addEventListener('click', () => map.zoomIn());
    document.getElementById('zoomOut').addEventListener('click', () => map.zoomOut());
    document.getElementById('centerMap').addEventListener('click', centerMapOnData);
    
    // Rating slider
    const scoreSlider = document.getElementById('scoreMin');
    scoreSlider.addEventListener('input', (e) => {
        document.getElementById('scoreMinValue').textContent = e.target.value;
    });
    
    // View mode
    document.getElementById('viewMode').addEventListener('change', (e) => {
        updateViewMode(e.target.value);
    });
    
    // Search and sort
    document.getElementById('searchInput').addEventListener('input', filterPropertiesList);
    document.getElementById('sortBy').addEventListener('change', sortPropertiesList);
}

// Load initial data
async function loadInitialData() {
    showLoading(true);
    await Promise.all([
        loadCities(),
        loadStats(),
        loadHeatmapData()
    ]);
    showLoading(false);
}

// Load heatmap data
async function loadHeatmapData() {
    const bounds = map.getBounds();
    if (!bounds) return;
    
    const ne = bounds.getNorthEast();
    const sw = bounds.getSouthWest();
    
    const filters = getFilters();
    let url = `${API_BASE_URL}/heatmap?lat_min=${sw.lat}&lat_max=${ne.lat}&lng_min=${sw.lng}&lng_max=${ne.lng}`;
    
    if (filters.city) url += `&city=${encodeURIComponent(filters.city)}`;
    if (filters.type) url += `&type=${encodeURIComponent(filters.type)}`;
    if (filters.priceMin) url += `&price_min=${filters.priceMin}`;
    if (filters.priceMax) url += `&price_max=${filters.priceMax}`;
    
    try {
        showLoading(true);
        const response = await fetch(url);
        
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        
        const data = await response.json();
        
        if (data.data && data.data.length > 0) {
            updateHeatmap(data.data);
            propertiesData = data.data;
            updateStats(data.count);
        } else {
            showMessage('No data found for selected area', 'info');
        }
    } catch (error) {
        console.error('Error loading heatmap data:', error);
        showError('Error loading data. Check server connection.');
    } finally {
        showLoading(false);
    }
}

// Update heatmap (Leaflet Heat)
function updateHeatmap(data) {
    const viewMode = document.getElementById('viewMode').value;
    
    if (viewMode === 'heatmap' || viewMode === 'both') {
        updateHeatmapLayer(data);
    }
    
    if (viewMode === 'markers' || viewMode === 'both') {
        updateMarkers(data);
    }
}

// Update heatmap layer
function updateHeatmapLayer(data) {
    // Remove old layer
    if (heatmapLayer) {
        map.removeLayer(heatmapLayer);
    }
    
    // Create heatmap data
    const heatData = data.map(point => [
        point.lat,
        point.lng,
        calculateWeight(point)
    ]);
    
    if (heatData.length > 0) {
        heatmapLayer = L.heatLayer(heatData, {
            radius: getRadiusByZoom(),
            blur: 15,
            maxZoom: 17,
            gradient: {
                0.0: 'blue',
                0.5: 'cyan',
                0.7: 'lime',
                0.8: 'yellow',
                1.0: 'red'
            }
        }).addTo(map);
    }
}

// Update markers
function updateMarkers(data) {
    // Clear old markers
    markers.forEach(marker => map.removeLayer(marker));
    markers = [];
    
    // Create markers
    data.forEach(point => {
        const marker = createMarker(point);
        markers.push(marker);
    });
}

// Create marker
function createMarker(point) {
    const color = getColorByScore(point.score || 0);
    
    const marker = L.circleMarker([point.lat, point.lng], {
        radius: 8,
        fillColor: color,
        color: '#fff',
        weight: 2,
        opacity: 0.8,
        fillOpacity: 0.8
    }).addTo(map);
    
    const popup = L.popup().setContent(createInfoWindowContent(point));
    marker.bindPopup(popup);
    
    marker.on('click', () => {
        showDetailedInfo(point);
    });
    
    return marker;
}

// Helper functions
function getRadiusByZoom() {
    const zoom = map.getZoom();
    if (zoom < 10) return 20;
    if (zoom < 13) return 30;
    if (zoom < 15) return 50;
    return 80;
}

function calculateWeight(point) {
    const priceWeight = Math.log10(point.price + 1) / 10;
    const scoreWeight = (point.score || 0) / 100;
    return priceWeight * 0.6 + scoreWeight * 0.4;
}

function getColorByScore(score) {
    if (score >= 70) return '#2ecc71';
    if (score >= 50) return '#f39c12';
    return '#e74c3c';
}

function formatPrice(price) {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD',
        maximumFractionDigits: 0
    }).format(price);
}

function formatScore(score) {
    const badgeClass = score >= 70 ? 'score-high' : score >= 50 ? 'score-medium' : 'score-low';
    return `<span class="score-badge ${badgeClass}">${score.toFixed(1)}</span>`;
}

function createInfoWindowContent(point) {
    return `
        <div style="padding: 10px; min-width: 200px;">
            <h4 style="margin: 0 0 10px 0; color: #2c3e50;">Area Information</h4>
            <p style="margin: 5px 0;"><strong>Average price:</strong> ${formatPrice(point.price)}</p>
            <p style="margin: 5px 0;"><strong>Properties:</strong> ${point.count || 1}</p>
            <p style="margin: 5px 0;"><strong>Overall rating:</strong> ${formatScore(point.score || 0)}</p>
        </div>
    `;
}

// Load properties list
async function loadPropertiesList() {
    const filters = getFilters();
    let url = `${API_BASE_URL}/properties?limit=100`;
    
    if (filters.city) url += `&city=${encodeURIComponent(filters.city)}`;
    if (filters.type) url += `&type=${encodeURIComponent(filters.type)}`;
    if (filters.priceMin) url += `&price_min=${filters.priceMin}`;
    if (filters.priceMax) url += `&price_max=${filters.priceMax}`;
    
    try {
        const response = await fetch(url);
        const data = await response.json();
        
        if (data.data) {
            displayPropertiesList(data.data);
        }
    } catch (error) {
        console.error('Error loading properties:', error);
    }
}

// Show detailed information
async function showDetailedInfo(point) {
    currentSelectedPoint = point;
    
    const content = document.getElementById('infoContent');
    content.innerHTML = `
        <div class="info-item">
            <strong>Average price</strong>
            <span class="price">${formatPrice(point.price)}</span>
        </div>
        <div class="info-item">
            <strong>Number of properties</strong>
            <span>${point.count}</span>
        </div>
        <div class="info-item">
            <strong>Overall rating</strong>
            ${formatScore(point.score || 0)}
        </div>
        <div class="info-item">
            <strong>Coordinates</strong>
            <span>${point.lat.toFixed(6)}, ${point.lng.toFixed(6)}</span>
        </div>
    `;
    
    // Load detailed factor data
    await loadFactorsData(point);
    
    // Show panel
    document.getElementById('infoPanel').style.display = 'block';
    document.getElementById('factorsPanel').style.display = 'block';
}

// Load factors data
async function loadFactorsData(point) {
    try {
        const response = await fetch(
            `${API_BASE_URL}/properties?lat_min=${point.lat - 0.01}&lat_max=${point.lat + 0.01}&lng_min=${point.lng - 0.01}&lng_max=${point.lng + 0.01}&limit=10`
        );
        
        const data = await response.json();
        
        if (data.data && data.data.length > 0) {
            const avgFactors = calculateAverageFactors(data.data);
            displayFactorsChart(avgFactors);
            displayFactorsDetails(avgFactors);
        }
    } catch (error) {
        console.error('Error loading factors:', error);
    }
}

// Calculate average factors
function calculateAverageFactors(properties) {
    const factors = {
        crime: 0,
        transport: 0,
        education: 0,
        infrastructure: 0,
        overall: 0
    };
    
    let count = 0;
    
    properties.forEach(prop => {
        if (prop.factors) {
            factors.crime += prop.factors.crime_score || 0;
            factors.transport += prop.factors.transport_score || 0;
            factors.education += prop.factors.education_score || 0;
            factors.infrastructure += prop.factors.infrastructure_score || 0;
            factors.overall += prop.factors.overall_score || 0;
            count++;
        }
    });
    
    if (count > 0) {
        Object.keys(factors).forEach(key => {
            factors[key] = factors[key] / count;
        });
    }
    
    return factors;
}

// Display factors chart
function displayFactorsChart(factors) {
    const ctx = document.getElementById('factorsChart');
    if (!ctx) return;
    
    if (factorsChart) {
        factorsChart.destroy();
    }
    
    factorsChart = new Chart(ctx, {
        type: 'radar',
        data: {
            labels: ['Safety', 'Transport', 'Education', 'Infrastructure'],
            datasets: [{
                label: 'Factor Rating',
                data: [
                    factors.crime,
                    factors.transport,
                    factors.education,
                    factors.infrastructure
                ],
                backgroundColor: 'rgba(102, 126, 234, 0.2)',
                borderColor: 'rgba(102, 126, 234, 1)',
                borderWidth: 2,
                pointBackgroundColor: 'rgba(102, 126, 234, 1)',
                pointBorderColor: '#fff',
                pointHoverBackgroundColor: '#fff',
                pointHoverBorderColor: 'rgba(102, 126, 234, 1)'
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                r: {
                    beginAtZero: true,
                    max: 100,
                    ticks: {
                        stepSize: 20
                    }
                }
            },
            plugins: {
                legend: {
                    display: false
                }
            }
        }
    });
}

// Display factors details
function displayFactorsDetails(factors) {
    const container = document.getElementById('factorsDetails');
    
    container.innerHTML = `
        <div class="factor-item" style="border-left: 4px solid #2ecc71;">
            <div class="factor-label">Safety</div>
            <div class="factor-value">${factors.crime.toFixed(1)}</div>
            <div class="progress-bar">
                <div class="progress-fill" style="width: ${factors.crime}%"></div>
            </div>
        </div>
        <div class="factor-item" style="border-left: 4px solid #3498db;">
            <div class="factor-label">Transport</div>
            <div class="factor-value">${factors.transport.toFixed(1)}</div>
            <div class="progress-bar">
                <div class="progress-fill" style="width: ${factors.transport}%"></div>
            </div>
        </div>
        <div class="factor-item" style="border-left: 4px solid #f39c12;">
            <div class="factor-label">Education</div>
            <div class="factor-value">${factors.education.toFixed(1)}</div>
            <div class="progress-bar">
                <div class="progress-fill" style="width: ${factors.education}%"></div>
            </div>
        </div>
        <div class="factor-item" style="border-left: 4px solid #9b59b6;">
            <div class="factor-label">Infrastructure</div>
            <div class="factor-value">${factors.infrastructure.toFixed(1)}</div>
            <div class="progress-bar">
                <div class="progress-fill" style="width: ${factors.infrastructure}%"></div>
            </div>
        </div>
    `;
}

// Display properties list
function displayPropertiesList(properties) {
    const container = document.getElementById('propertiesContainer');
    
    if (properties.length === 0) {
        container.innerHTML = '<p class="placeholder">No properties found</p>';
        return;
    }
    
    container.innerHTML = properties.map(prop => `
        <div class="property-card" onclick="focusOnProperty(${prop.latitude}, ${prop.longitude})">
            <h4>${prop.address || 'Address not specified'}</h4>
            <div class="price">${formatPrice(prop.price)}</div>
            <div class="details">
                ${prop.area ? `Area: ${prop.area} m²` : ''} | 
                ${prop.rooms ? `Rooms: ${prop.rooms}` : ''} | 
                ${prop.type || ''}
            </div>
            <div>Rating: ${formatScore(prop.factors?.overall_score || 0)}</div>
        </div>
    `).join('');
}

// Focus on property
function focusOnProperty(lat, lng) {
    map.setView([lat, lng], 15);
}

// Apply filters
function applyFilters() {
    loadHeatmapData();
    loadPropertiesList();
}

// Reset filters
function resetFilters() {
    document.getElementById('cityFilter').value = '';
    document.getElementById('typeFilter').value = '';
    document.getElementById('priceMin').value = '';
    document.getElementById('priceMax').value = '';
    document.getElementById('scoreMin').value = 0;
    document.getElementById('scoreMinValue').textContent = '0';
    applyFilters();
}

// Toggle auto-update
function toggleAutoUpdate() {
    const btn = document.getElementById('autoUpdate');
    const isActive = btn.classList.contains('active');
    
    if (isActive) {
        stopAutoUpdate();
        btn.classList.remove('active');
        btn.textContent = 'Auto-update: OFF';
    } else {
        startAutoUpdate();
        btn.classList.add('active');
        btn.textContent = 'Auto-update: ON';
    }
}

// Start auto-update
function startAutoUpdate() {
    if (autoUpdateInterval) return;
    
    autoUpdateInterval = setInterval(() => {
        loadHeatmapData();
        loadStats();
    }, 60000); // Every minute
}

// Stop auto-update
function stopAutoUpdate() {
    if (autoUpdateInterval) {
        clearInterval(autoUpdateInterval);
        autoUpdateInterval = null;
    }
}

// Toggle details panel
function toggleDetailsPanel() {
    const list = document.getElementById('propertiesList');
    const isVisible = list.style.display !== 'none';
    
    if (isVisible) {
        list.style.display = 'none';
        document.getElementById('toggleDetails').textContent = 'Show details';
    } else {
        list.style.display = 'block';
        document.getElementById('toggleDetails').textContent = 'Hide details';
        loadPropertiesList();
    }
}

// Close info panel
function closeInfoPanel() {
    document.getElementById('infoPanel').style.display = 'none';
    document.getElementById('factorsPanel').style.display = 'none';
    currentSelectedPoint = null;
}

// Center map on data
function centerMapOnData() {
    if (propertiesData.length === 0) return;
    
    const bounds = L.latLngBounds(propertiesData.map(p => [p.lat, p.lng]));
    map.fitBounds(bounds);
}

// Update view mode
function updateViewMode(mode) {
    if (mode === 'heatmap') {
        markers.forEach(m => map.removeLayer(m));
        if (heatmapLayer) map.addLayer(heatmapLayer);
    } else if (mode === 'markers') {
        if (heatmapLayer) map.removeLayer(heatmapLayer);
        markers.forEach(m => map.addLayer(m));
    } else {
        if (heatmapLayer) map.addLayer(heatmapLayer);
        markers.forEach(m => map.addLayer(m));
    }
}

// Filter properties list
function filterPropertiesList() {
    const search = document.getElementById('searchInput').value.toLowerCase();
    const cards = document.querySelectorAll('.property-card');
    
    cards.forEach(card => {
        const text = card.textContent.toLowerCase();
        card.style.display = text.includes(search) ? 'block' : 'none';
    });
}

// Sort properties list
function sortPropertiesList() {
    loadPropertiesList();
}

// Load cities
async function loadCities() {
    try {
        const response = await fetch(`${API_BASE_URL}/stats`);
        const data = await response.json();
        
        const citySelect = document.getElementById('cityFilter');
        citySelect.innerHTML = '<option value="">All cities</option>';
        
        if (data.cities) {
            data.cities.forEach(city => {
                const option = document.createElement('option');
                option.value = city;
                option.textContent = city;
                citySelect.appendChild(option);
            });
        }
    } catch (error) {
        console.error('Error loading cities:', error);
    }
}

// Load stats
async function loadStats() {
    try {
        const response = await fetch(`${API_BASE_URL}/stats`);
        const data = await response.json();
        
        document.getElementById('totalProperties').textContent = data.total_properties || 0;
        document.getElementById('avgPrice').textContent = formatPrice(data.avg_price || 0);
        updateLastUpdateTime();
    } catch (error) {
        console.error('Error loading stats:', error);
    }
}

// Update stats
function updateStats(count) {
    document.getElementById('totalProperties').textContent = count;
    updateLastUpdateTime();
}

// Update last update time
function updateLastUpdateTime() {
    const now = new Date();
    const time = now.toLocaleTimeString('en-US');
    document.getElementById('lastUpdate').textContent = time;
}

// Get filters
function getFilters() {
    return {
        city: document.getElementById('cityFilter').value,
        type: document.getElementById('typeFilter').value,
        priceMin: document.getElementById('priceMin').value,
        priceMax: document.getElementById('priceMax').value,
        scoreMin: document.getElementById('scoreMin').value
    };
}

// Show loading
function showLoading(show) {
    const overlay = document.getElementById('loadingOverlay');
    if (show) {
        overlay.classList.add('active');
    } else {
        overlay.classList.remove('active');
    }
}

// Show error
function showError(message) {
    showMessage(message, 'error');
}

// Show message
function showMessage(message, type = 'info') {
    console.log(`[${type.toUpperCase()}] ${message}`);
}

// Export function for use in HTML
window.focusOnProperty = focusOnProperty;
