// Конфигурация
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

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    initApp();
});

// Инициализация приложения
function initApp() {
    if (typeof google === 'undefined' || !google.maps) {
        console.error('Google Maps API не загружена');
        showError('Ошибка загрузки Google Maps API. Проверьте API ключ.');
        return;
    }
    
    initMap();
    initEventListeners();
    loadInitialData();
    startAutoUpdate();
}

// Инициализация карты
function initMap() {
    map = new google.maps.Map(document.getElementById('map'), {
        center: { lat: 55.7558, lng: 37.6173 }, // Москва по умолчанию
        zoom: 10,
        mapTypeId: 'roadmap',
        styles: [
            {
                featureType: 'poi',
                elementType: 'labels',
                stylers: [{ visibility: 'off' }]
            }
        ]
    });
    
    // Debounced загрузка данных при изменении границ
    map.addListener('bounds_changed', () => {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(() => {
            loadHeatmapData();
        }, 500);
    });
    
    // Загрузка данных при изменении zoom
    map.addListener('zoom_changed', () => {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(() => {
            loadHeatmapData();
        }, 300);
    });
}

// Инициализация обработчиков событий
function initEventListeners() {
    // Фильтры
    document.getElementById('applyFilters').addEventListener('click', applyFilters);
    document.getElementById('resetFilters').addEventListener('click', resetFilters);
    
    // Автообновление
    document.getElementById('autoUpdate').addEventListener('click', toggleAutoUpdate);
    
    // Детали
    document.getElementById('toggleDetails').addEventListener('click', toggleDetailsPanel);
    document.getElementById('closePanel').addEventListener('click', closeInfoPanel);
    
    // Управление картой
    document.getElementById('zoomIn').addEventListener('click', () => map.setZoom(map.getZoom() + 1));
    document.getElementById('zoomOut').addEventListener('click', () => map.setZoom(map.getZoom() - 1));
    document.getElementById('centerMap').addEventListener('click', centerMapOnData);
    
    // Слайдер рейтинга
    const scoreSlider = document.getElementById('scoreMin');
    scoreSlider.addEventListener('input', (e) => {
        document.getElementById('scoreMinValue').textContent = e.target.value;
    });
    
    // Режим отображения
    document.getElementById('viewMode').addEventListener('change', (e) => {
        updateViewMode(e.target.value);
    });
    
    // Поиск и сортировка
    document.getElementById('searchInput').addEventListener('input', filterPropertiesList);
    document.getElementById('sortBy').addEventListener('change', sortPropertiesList);
}

// Загрузка начальных данных
async function loadInitialData() {
    showLoading(true);
    await Promise.all([
        loadCities(),
        loadStats(),
        loadHeatmapData()
    ]);
    showLoading(false);
}

// Загрузка данных для heatmap
async function loadHeatmapData() {
    const bounds = map.getBounds();
    if (!bounds) return;
    
    const ne = bounds.getNorthEast();
    const sw = bounds.getSouthWest();
    
    const filters = getFilters();
    let url = `${API_BASE_URL}/heatmap?lat_min=${sw.lat()}&lat_max=${ne.lat()}&lng_min=${sw.lng()}&lng_max=${ne.lng()}`;
    
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
            showMessage('Данные не найдены для выбранной области', 'info');
        }
    } catch (error) {
        console.error('Error loading heatmap data:', error);
        showError('Ошибка загрузки данных. Проверьте подключение к серверу.');
    } finally {
        showLoading(false);
    }
}

// Загрузка списка объектов
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

// Обновление heatmap
function updateHeatmap(data) {
    const viewMode = document.getElementById('viewMode').value;
    
    if (viewMode === 'heatmap' || viewMode === 'both') {
        updateHeatmapLayer(data);
    }
    
    if (viewMode === 'markers' || viewMode === 'both') {
        updateMarkers(data);
    }
}

// Обновление heatmap слоя
function updateHeatmapLayer(data) {
    heatmapData = data.map(point => ({
        location: new google.maps.LatLng(point.lat, point.lng),
        weight: calculateWeight(point)
    }));
    
    if (heatmapLayer) {
        heatmapLayer.setMap(null);
    }
    
    if (heatmapData.length > 0) {
        heatmapLayer = new google.maps.visualization.HeatmapLayer({
            data: heatmapData,
            map: map,
            radius: getRadiusByZoom(),
            opacity: 0.7,
            gradient: getGradient(),
            maxIntensity: getMaxIntensity(data)
        });
    }
}

// Обновление маркеров
function updateMarkers(data) {
    // Очищаем старые маркеры
    markers.forEach(marker => marker.setMap(null));
    markers = [];
    
    // Создаем кластеры для больших наборов данных
    if (data.length > 100) {
        createClusteredMarkers(data);
    } else {
        createIndividualMarkers(data);
    }
}

// Создание индивидуальных маркеров
function createIndividualMarkers(data) {
    data.forEach(point => {
        const marker = createMarker(point);
        markers.push(marker);
    });
}

// Создание кластеризованных маркеров
function createClusteredMarkers(data) {
    // Упрощенная кластеризация
    const gridSize = 0.01;
    const clusters = {};
    
    data.forEach(point => {
        const lat = Math.round(point.lat / gridSize) * gridSize;
        const lng = Math.round(point.lng / gridSize) * gridSize;
        const key = `${lat},${lng}`;
        
        if (!clusters[key]) {
            clusters[key] = {
                lat, lng,
                points: [],
                avgPrice: 0,
                avgScore: 0
            };
        }
        
        clusters[key].points.push(point);
        clusters[key].avgPrice += point.price;
        clusters[key].avgScore += point.score || 0;
    });
    
    Object.values(clusters).forEach(cluster => {
        cluster.avgPrice /= cluster.points.length;
        cluster.avgScore /= cluster.points.length;
        
        const marker = createMarker({
            lat: cluster.lat,
            lng: cluster.lng,
            price: cluster.avgPrice,
            score: cluster.avgScore,
            count: cluster.points.length
        }, true);
        
        markers.push(marker);
    });
}

// Создание маркера
function createMarker(point, isCluster = false) {
    const size = isCluster ? Math.min(point.count * 2, 15) : 8;
    
    const marker = new google.maps.Marker({
        position: { lat: point.lat, lng: point.lng },
        map: map,
        title: `Цена: ${formatPrice(point.price)}, Рейтинг: ${point.score?.toFixed(1) || 'N/A'}`,
        icon: {
            path: google.maps.SymbolPath.CIRCLE,
            scale: size,
            fillColor: getColorByScore(point.score || 0),
            fillOpacity: 0.8,
            strokeColor: '#fff',
            strokeWeight: 2
        },
        animation: google.maps.Animation.DROP
    });
    
    const infoWindow = new google.maps.InfoWindow({
        content: createInfoWindowContent(point, isCluster)
    });
    
    marker.addListener('click', () => {
        infoWindow.open(map, marker);
        showDetailedInfo(point);
        highlightMarker(marker);
    });
    
    return marker;
}

// Подсветка маркера
function highlightMarker(marker) {
    if (marker.getAnimation() !== null) {
        marker.setAnimation(null);
    } else {
        marker.setAnimation(google.maps.Animation.BOUNCE);
        setTimeout(() => marker.setAnimation(null), 2000);
    }
}

// Обновление режима отображения
function updateViewMode(mode) {
    if (mode === 'heatmap') {
        markers.forEach(m => m.setMap(null));
        if (heatmapLayer) heatmapLayer.setMap(map);
    } else if (mode === 'markers') {
        if (heatmapLayer) heatmapLayer.setMap(null);
        markers.forEach(m => m.setMap(map));
    } else {
        if (heatmapLayer) heatmapLayer.setMap(map);
        markers.forEach(m => m.setMap(map));
    }
}

// Вычисление веса для heatmap
function calculateWeight(point) {
    const priceWeight = Math.log10(point.price + 1) / 10;
    const scoreWeight = (point.score || 0) / 100;
    return priceWeight * 0.6 + scoreWeight * 0.4;
}

// Получение радиуса по zoom
function getRadiusByZoom() {
    const zoom = map.getZoom();
    if (zoom < 10) return 30;
    if (zoom < 13) return 50;
    if (zoom < 15) return 80;
    return 100;
}

// Получение градиента
function getGradient() {
    return [
        'rgba(0, 255, 255, 0)',
        'rgba(0, 255, 255, 1)',
        'rgba(0, 191, 255, 1)',
        'rgba(0, 127, 255, 1)',
        'rgba(0, 63, 255, 1)',
        'rgba(0, 0, 255, 1)',
        'rgba(0, 0, 223, 1)',
        'rgba(0, 0, 191, 1)',
        'rgba(0, 0, 159, 1)',
        'rgba(0, 0, 127, 1)',
        'rgba(63, 0, 91, 1)',
        'rgba(127, 0, 63, 1)',
        'rgba(191, 0, 31, 1)',
        'rgba(255, 0, 0, 1)'
    ];
}

// Получение максимальной интенсивности
function getMaxIntensity(data) {
    if (data.length === 0) return 1;
    const maxWeight = Math.max(...data.map(p => calculateWeight(p)));
    return maxWeight * 1.2;
}

// Показ детальной информации
async function showDetailedInfo(point) {
    currentSelectedPoint = point;
    
    const content = document.getElementById('infoContent');
    content.innerHTML = `
        <div class="info-item">
            <strong>Средняя цена</strong>
            <span class="price">${formatPrice(point.price)}</span>
        </div>
        <div class="info-item">
            <strong>Количество объектов</strong>
            <span>${point.count}</span>
        </div>
        <div class="info-item">
            <strong>Общий рейтинг</strong>
            ${formatScore(point.score || 0)}
        </div>
        <div class="info-item">
            <strong>Координаты</strong>
            <span>${point.lat.toFixed(6)}, ${point.lng.toFixed(6)}</span>
        </div>
    `;
    
    // Загружаем детальные данные о факторах
    await loadFactorsData(point);
    
    // Показываем панель
    document.getElementById('infoPanel').style.display = 'block';
    document.getElementById('factorsPanel').style.display = 'block';
}

// Загрузка данных о факторах
async function loadFactorsData(point) {
    try {
        // Ищем ближайшие объекты для получения детальных факторов
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

// Вычисление средних факторов
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

// Отображение графика факторов
function displayFactorsChart(factors) {
    const ctx = document.getElementById('factorsChart');
    if (!ctx) return;
    
    if (factorsChart) {
        factorsChart.destroy();
    }
    
    factorsChart = new Chart(ctx, {
        type: 'radar',
        data: {
            labels: ['Безопасность', 'Транспорт', 'Образование', 'Инфраструктура'],
            datasets: [{
                label: 'Оценка факторов',
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

// Отображение деталей факторов
function displayFactorsDetails(factors) {
    const container = document.getElementById('factorsDetails');
    
    container.innerHTML = `
        <div class="factor-item" style="border-left: 4px solid #2ecc71;">
            <div class="factor-label">Безопасность</div>
            <div class="factor-value">${factors.crime.toFixed(1)}</div>
            <div class="progress-bar">
                <div class="progress-fill" style="width: ${factors.crime}%"></div>
            </div>
        </div>
        <div class="factor-item" style="border-left: 4px solid #3498db;">
            <div class="factor-label">Транспорт</div>
            <div class="factor-value">${factors.transport.toFixed(1)}</div>
            <div class="progress-bar">
                <div class="progress-fill" style="width: ${factors.transport}%"></div>
            </div>
        </div>
        <div class="factor-item" style="border-left: 4px solid #f39c12;">
            <div class="factor-label">Образование</div>
            <div class="factor-value">${factors.education.toFixed(1)}</div>
            <div class="progress-bar">
                <div class="progress-fill" style="width: ${factors.education}%"></div>
            </div>
        </div>
        <div class="factor-item" style="border-left: 4px solid #9b59b6;">
            <div class="factor-label">Инфраструктура</div>
            <div class="factor-value">${factors.infrastructure.toFixed(1)}</div>
            <div class="progress-bar">
                <div class="progress-fill" style="width: ${factors.infrastructure}%"></div>
            </div>
        </div>
    `;
}

// Отображение списка объектов
function displayPropertiesList(properties) {
    const container = document.getElementById('propertiesContainer');
    
    if (properties.length === 0) {
        container.innerHTML = '<p class="placeholder">Объекты не найдены</p>';
        return;
    }
    
    container.innerHTML = properties.map(prop => `
        <div class="property-card" onclick="focusOnProperty(${prop.latitude}, ${prop.longitude})">
            <h4>${prop.address || 'Адрес не указан'}</h4>
            <div class="price">${formatPrice(prop.price)}</div>
            <div class="details">
                ${prop.area ? `Площадь: ${prop.area} м²` : ''} | 
                ${prop.rooms ? `Комнат: ${prop.rooms}` : ''} | 
                ${prop.type || ''}
            </div>
            <div>Рейтинг: ${formatScore(prop.factors?.overall_score || 0)}</div>
        </div>
    `).join('');
}

// Фокус на объекте
function focusOnProperty(lat, lng) {
    map.setCenter({ lat, lng });
    map.setZoom(15);
}

// Фильтрация списка объектов
function filterPropertiesList() {
    const search = document.getElementById('searchInput').value.toLowerCase();
    const cards = document.querySelectorAll('.property-card');
    
    cards.forEach(card => {
        const text = card.textContent.toLowerCase();
        card.style.display = text.includes(search) ? 'block' : 'none';
    });
}

// Сортировка списка объектов
function sortPropertiesList() {
    // Реализация сортировки через API
    loadPropertiesList();
}

// Загрузка статистики
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

// Обновление статистики
function updateStats(count) {
    document.getElementById('totalProperties').textContent = count;
    updateLastUpdateTime();
}

// Обновление времени последнего обновления
function updateLastUpdateTime() {
    const now = new Date();
    const time = now.toLocaleTimeString('ru-RU');
    document.getElementById('lastUpdate').textContent = time;
}

// Загрузка списка городов
async function loadCities() {
    try {
        const response = await fetch(`${API_BASE_URL}/stats`);
        const data = await response.json();
        
        const citySelect = document.getElementById('cityFilter');
        citySelect.innerHTML = '<option value="">Все города</option>';
        
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

// Получение фильтров
function getFilters() {
    return {
        city: document.getElementById('cityFilter').value,
        type: document.getElementById('typeFilter').value,
        priceMin: document.getElementById('priceMin').value,
        priceMax: document.getElementById('priceMax').value,
        scoreMin: document.getElementById('scoreMin').value
    };
}

// Применение фильтров
function applyFilters() {
    loadHeatmapData();
    loadPropertiesList();
}

// Сброс фильтров
function resetFilters() {
    document.getElementById('cityFilter').value = '';
    document.getElementById('typeFilter').value = '';
    document.getElementById('priceMin').value = '';
    document.getElementById('priceMax').value = '';
    document.getElementById('scoreMin').value = 0;
    document.getElementById('scoreMinValue').textContent = '0';
    
    applyFilters();
}

// Переключение автообновления
function toggleAutoUpdate() {
    const btn = document.getElementById('autoUpdate');
    const isActive = btn.classList.contains('active');
    
    if (isActive) {
        stopAutoUpdate();
        btn.classList.remove('active');
        btn.textContent = 'Автообновление: ВЫКЛ';
    } else {
        startAutoUpdate();
        btn.classList.add('active');
        btn.textContent = 'Автообновление: ВКЛ';
    }
}

// Запуск автообновления
function startAutoUpdate() {
    if (autoUpdateInterval) return;
    
    autoUpdateInterval = setInterval(() => {
        loadHeatmapData();
        loadStats();
    }, 60000); // Каждую минуту
}

// Остановка автообновления
function stopAutoUpdate() {
    if (autoUpdateInterval) {
        clearInterval(autoUpdateInterval);
        autoUpdateInterval = null;
    }
}

// Переключение панели деталей
function toggleDetailsPanel() {
    const list = document.getElementById('propertiesList');
    const isVisible = list.style.display !== 'none';
    
    if (isVisible) {
        list.style.display = 'none';
        document.getElementById('toggleDetails').textContent = 'Показать детали';
    } else {
        list.style.display = 'block';
        document.getElementById('toggleDetails').textContent = 'Скрыть детали';
        loadPropertiesList();
    }
}

// Закрытие панели информации
function closeInfoPanel() {
    document.getElementById('infoPanel').style.display = 'none';
    document.getElementById('factorsPanel').style.display = 'none';
    currentSelectedPoint = null;
}

// Центрирование карты на данных
function centerMapOnData() {
    if (propertiesData.length === 0) return;
    
    const bounds = new google.maps.LatLngBounds();
    propertiesData.forEach(point => {
        bounds.extend(new google.maps.LatLng(point.lat, point.lng));
    });
    
    map.fitBounds(bounds);
}

// Показ загрузки
function showLoading(show) {
    const overlay = document.getElementById('loadingOverlay');
    if (show) {
        overlay.classList.add('active');
    } else {
        overlay.classList.remove('active');
    }
}

// Показ ошибки
function showError(message) {
    showMessage(message, 'error');
}

// Показ сообщения
function showMessage(message, type = 'info') {
    // Простая реализация - можно улучшить
    console.log(`[${type.toUpperCase()}] ${message}`);
}

// Создание контента для InfoWindow
function createInfoWindowContent(point, isCluster = false) {
    return `
        <div style="padding: 10px; min-width: 200px;">
            <h4 style="margin: 0 0 10px 0; color: #2c3e50;">${isCluster ? 'Кластер объектов' : 'Информация о районе'}</h4>
            <p style="margin: 5px 0;"><strong>Средняя цена:</strong> ${formatPrice(point.price)}</p>
            <p style="margin: 5px 0;"><strong>Объектов:</strong> ${point.count || 1}</p>
            <p style="margin: 5px 0;"><strong>Общий рейтинг:</strong> ${formatScore(point.score || 0)}</p>
            ${isCluster ? '<p style="margin: 5px 0; font-size: 0.9em; color: #666;">Кликните для деталей</p>' : ''}
        </div>
    `;
}

// Форматирование цены
function formatPrice(price) {
    return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency: 'USD',
        maximumFractionDigits: 0
    }).format(price);
}

// Форматирование рейтинга
function formatScore(score) {
    const badgeClass = score >= 70 ? 'score-high' : score >= 50 ? 'score-medium' : 'score-low';
    return `<span class="score-badge ${badgeClass}">${score.toFixed(1)}</span>`;
}

// Получение цвета по рейтингу
function getColorByScore(score) {
    if (score >= 70) return '#2ecc71';
    if (score >= 50) return '#f39c12';
    return '#e74c3c';
}

// Экспорт функции для использования в HTML
window.focusOnProperty = focusOnProperty;
