// Конфигурация
const API_BASE_URL = 'http://localhost:8000';

// Элементы DOM
const numberInput = document.getElementById('numberInput');
const predictBtn = document.getElementById('predictBtn');
const trainBtn = document.getElementById('trainBtn');
const clearBtn = document.getElementById('clearBtn');
const resultDiv = document.getElementById('result');
const loadingDiv = document.getElementById('loading');
const errorDiv = document.getElementById('error');
const errorText = document.getElementById('errorText');
const apiStatus = document.getElementById('apiStatus');

// Элементы для отображения результата
const confidenceValue = document.getElementById('confidenceValue');
const predictedPrice = document.getElementById('predictedPrice');
const rangeText = document.getElementById('rangeText');
const minPrice = document.getElementById('minPrice');
const maxPrice = document.getElementById('maxPrice');
const minPriceDetail = document.getElementById('minPriceDetail');
const maxPriceDetail = document.getElementById('maxPriceDetail');
const confidenceDetail = document.getElementById('confidenceDetail');
const rangeMin = document.getElementById('rangeMin');
const rangeMax = document.getElementById('rangeMax');
const rangePredicted = document.getElementById('rangePredicted');
const predictedTooltip = document.getElementById('predictedTooltip');

// Элементы номера
const letter1 = document.getElementById('letter1');
const letter2 = document.getElementById('letter2');
const letter3 = document.getElementById('letter3');
const digit1 = document.getElementById('digit1');
const digit2 = document.getElementById('digit2');
const digit3 = document.getElementById('digit3');
const region = document.getElementById('region');

// Модальное окно
const trainModal = document.getElementById('trainModal');
const confirmTrainBtn = document.getElementById('confirmTrainBtn');
const cancelTrainBtn = document.getElementById('cancelTrainBtn');

// История
const historyDiv = document.getElementById('history');
const historyList = document.getElementById('historyList');
const clearHistoryBtn = document.getElementById('clearHistoryBtn');
const historyBtn = document.getElementById('historyBtn');

// Инициализация
document.addEventListener('DOMContentLoaded', () => {
    loadHistory();
    checkAPIStatus();

    // Автопроверка статуса API каждые 30 секунд
    setInterval(checkAPIStatus, 30000);
});

// Проверка статуса API
async function checkAPIStatus() {
    try {
        const response = await fetch(`${API_BASE_URL}/`);
        if (response.ok) {
            apiStatus.textContent = 'онлайн';
            apiStatus.className = 'status-online';
        } else {
            apiStatus.textContent = 'ошибка';
            apiStatus.className = 'status-offline';
        }
    } catch (error) {
        apiStatus.textContent = 'оффлайн';
        apiStatus.className = 'status-offline';
    }
}

// Валидация и форматирование номера
function formatNumber(input) {
    // Удаляем все пробелы и приводим к верхнему регистру
    let value = input.replace(/\s+/g, '').toUpperCase();

    // Оставляем только разрешенные символы
    value = value.replace(/[^АВЕКМНОРСТУХ0123456789]/g, '');

    return value;
}

function isValidNumber(number) {
    // Проверяем формат: буква + 3 цифры + 2 буквы + 2-3 цифры
    const pattern = /^[АВЕКМНОРСТУХ]\d{3}[АВЕКМНОРСТУХ]{2}\d{2,3}$/;
    return pattern.test(number);
}

// Разбор номера на части для отображения
function parseNumber(number) {
    if (!isValidNumber(number)) return null;

    return {
        letter1: number[0],
        digits: number.slice(1, 4),
        letter2: number[4],
        letter3: number[5],
        region: number.slice(6)
    };
}

// Отображение номера по частям
function displayNumber(number) {
    const parts = parseNumber(number);
    if (!parts) return;

    letter1.textContent = parts.letter1;
    letter2.textContent = parts.letter2;
    letter3.textContent = parts.letter3;
    digit1.textContent = parts.digits[0];
    digit2.textContent = parts.digits[1];
    digit3.textContent = parts.digits[2];
    region.textContent = parts.region;
}

// Форматирование цены
function formatPrice(price) {
    return new Intl.NumberFormat('ru-RU').format(price) + ' ₽';
}

// Обновление шкалы цен
function updatePriceRange(min, max, predicted) {
    // Преобразуем в числа
    min = parseInt(min);
    max = parseInt(max);
    predicted = parseInt(predicted);

    // Рассчитываем позиции на шкале (0-100%)
    const totalRange = max - min;
    const minPos = 0;
    const maxPos = 100;
    const predictedPos = totalRange > 0 ? ((predicted - min) / totalRange) * 100 : 50;

    // Устанавливаем позиции элементов
    rangeMin.style.left = `${minPos}%`;
    rangeMax.style.left = `${maxPos}%`;
    rangePredicted.style.left = `${predictedPos}%`;

    // Обновляем тултип
    predictedTooltip.textContent = formatPrice(predicted);
    predictedTooltip.style.left = `${predictedPos}%`;
}

// Показ результата
function showResult(data) {
    const predicted = data.predicted_price;
    const min = data.price_range.low;
    const max = data.price_range.high;
    const confidence = Math.round(data.confidence * 100);

    // Обновляем текстовые значения
    confidenceValue.textContent = `${confidence}%`;
    confidenceDetail.textContent = `${confidence}%`;
    predictedPrice.textContent = formatPrice(predicted);
    rangeText.textContent = `${formatPrice(min)} - ${formatPrice(max)}`;
    minPrice.textContent = formatPrice(min);
    maxPrice.textContent = formatPrice(max);
    minPriceDetail.textContent = formatPrice(min);
    maxPriceDetail.textContent = formatPrice(max);

    // Обновляем шкалу
    updatePriceRange(min, max, predicted);

    // Отображаем результат
    loadingDiv.classList.add('hidden');
    errorDiv.classList.add('hidden');
    resultDiv.classList.remove('hidden');

    // Сохраняем в историю
    saveToHistory(data);
}

// Сохранение в историю
function saveToHistory(data) {
    const history = getHistory();
    history.unshift({
        number: data.number,
        price: data.predicted_price,
        min: data.price_range.low,
        max: data.price_range.high,
        confidence: data.confidence,
        timestamp: new Date().toISOString()
    });

    // Ограничиваем историю 10 записями
    if (history.length > 10) {
        history.pop();
    }

    localStorage.setItem('numberPriceHistory', JSON.stringify(history));
    loadHistory();
}

// Получение истории
function getHistory() {
    const historyJson = localStorage.getItem('numberPriceHistory');
    return historyJson ? JSON.parse(historyJson) : [];
}

// Загрузка истории
function loadHistory() {
    const history = getHistory();
    historyList.innerHTML = '';

    if (history.length === 0) {
        historyList.innerHTML = '<div class="history-item">История пуста</div>';
        return;
    }

    history.forEach(item => {
        const div = document.createElement('div');
        div.className = 'history-item';
        div.innerHTML = `
            <span>${item.number}</span>
            <strong>${formatPrice(item.price)}</strong>
        `;
        div.onclick = () => {
            numberInput.value = item.number;
            predict();
        };
        historyList.appendChild(div);
    });
}

// Очистка истории
function clearHistory() {
    localStorage.removeItem('numberPriceHistory');
    loadHistory();
}

// Показ ошибки
function showError(message) {
    errorText.textContent = message;
    loadingDiv.classList.add('hidden');
    resultDiv.classList.add('hidden');
    errorDiv.classList.remove('hidden');
}

// Запрос на предсказание
async function predict() {
    let number = formatNumber(numberInput.value);

    if (!number) {
        showError('Введите номер автомобиля');
        return;
    }

    if (!isValidNumber(number)) {
        showError('Некорректный формат номера. Используйте формат: А123ВС777');
        return;
    }

    // Отображаем номер
    displayNumber(number);

    // Показываем загрузку
    loadingDiv.classList.remove('hidden');
    resultDiv.classList.add('hidden');
    errorDiv.classList.add('hidden');

    try {
        const response = await fetch(`${API_BASE_URL}/api/predict?number=${encodeURIComponent(number)}`);

        if (response.ok) {
            const data = await response.json();
            showResult(data);
        } else {
            const error = await response.text();
            showError(`Ошибка сервера: ${error}`);
        }
    } catch (error) {
        showError('Не удалось подключиться к серверу. Проверьте подключение к API.');
    }
}

// Запрос на обучение модели
async function trainModel(daysBack = 365) {
    try {
        const response = await fetch(`${API_BASE_URL}/api/train`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ days_back: parseInt(daysBack) })
        });

        if (response.ok) {
            alert('Обучение модели успешно запущено! Это может занять несколько минут.');
        } else {
            const error = await response.text();
            alert(`Ошибка при запуске обучения: ${error}`);
        }
    } catch (error) {
        alert('Не удалось подключиться к серверу для обучения модели.');
    }
}

// Обработчики событий
numberInput.addEventListener('input', (e) => {
    const formatted = formatNumber(e.target.value);
    e.target.value = formatted;

    // Авто-предсказание при вводе полного номера
    if (formatted.length >= 8 && isValidNumber(formatted)) {
        displayNumber(formatted);
    }
});

numberInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        predict();
    }
});

predictBtn.addEventListener('click', predict);

clearBtn.addEventListener('click', () => {
    numberInput.value = '';
    numberInput.focus();
});

trainBtn.addEventListener('click', () => {
    trainModal.classList.remove('hidden');
});

confirmTrainBtn.addEventListener('click', () => {
    const selectedDays = document.querySelector('input[name="days"]:checked').value;
    trainModal.classList.add('hidden');
    trainModel(selectedDays);
});

cancelTrainBtn.addEventListener('click', () => {
    trainModal.classList.add('hidden');
});

// Закрытие модального окна при клике вне его
window.addEventListener('click', (e) => {
    if (e.target === trainModal) {
        trainModal.classList.add('hidden');
    }
});

historyBtn.addEventListener('click', () => {
    historyDiv.classList.toggle('hidden');
    if (!historyDiv.classList.contains('hidden')) {
        loadHistory();
    }
});

clearHistoryBtn.addEventListener('click', clearHistory);

// Поддержка функции "Поделиться"
document.getElementById('shareBtn').addEventListener('click', async () => {
    const number = formatNumber(numberInput.value);
    if (!number) {
        alert('Введите номер для того, чтобы поделиться результатом');
        return;
    }

    const shareText = `Цена автономера ${number}: ${predictedPrice.textContent}. ${rangeText.textContent}`;

    if (navigator.share) {
        try {
            await navigator.share({
                title: 'Предсказание цены автономера',
                text: shareText,
                url: window.location.href
            });
        } catch (error) {
            console.log('Ошибка при использовании Web Share API:', error);
            copyToClipboard(shareText);
        }
    } else {
        copyToClipboard(shareText);
    }
});

// Копирование в буфер обмена
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        alert('Результат скопирован в буфер обмена!');
    });
}