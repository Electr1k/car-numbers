from fastapi import FastAPI, HTTPException, BackgroundTasks
from pydantic import BaseModel
import asyncio
from typing import Optional
import os

from data_loader import DataLoader
from price_predictor import NumberPricePredictor
from feature_extractor import FeatureEngineer

# Инициализация
predictor = None
is_training = False

# Pydantic модели
class TrainRequest(BaseModel):
    days_back: int = 365

# FastAPI приложение
app = FastAPI(title="Car Number Price API", version="1.0")

def init_predictor():
    """Загружаем модель при старте"""
    global predictor
    predictor = NumberPricePredictor()
    try:
        predictor.load_model()
        print("✅ Модель загружена")
    except:
        print("⚠️ Модель не загружена")

# Загружаем модель при импорте
init_predictor()

def train_model_sync(days_back: int):
    """Синхронное обучение модели"""
    global is_training, predictor
    
    try:
        is_training = True
        
        # Конфигурация БД
        db_config = {
            'host': os.getenv('DATABASE_HOST', 'postgres'),
            'database': os.getenv('DATABASE_NAME', 'postgres'),
            'user': os.getenv('DATABASE_USER', 'postgres'),
            'password': os.getenv('DATABASE_PASSWORD', 'postgres1'),
            'port': os.getenv('DATABASE_PORT', '5432')
        }
        
        # Загрузка данных
        loader = DataLoader(db_config)
        raw_data = loader.load_data(days_back=days_back)
        loader.close()
        
        # Подготовка признаков
        feature_engineer = FeatureEngineer()
        processed_data = feature_engineer.prepare_dataframe(raw_data)
        
        # 3. Обучение модели
        print("\nШаг 3: Обучение модели...")
        predictor = NumberPricePredictor()
        model = predictor.train(processed_data)

        print("✅ Обучение завершено")
        return {"success": True, "message": "Модель обучена"}
        
    except Exception as e:
        print(f"❌ Ошибка обучения: {e}")
        return {"success": False, "error": str(e)}
    finally:
        is_training = False

async def train_background(days_back: int):
    """Запуск обучения в фоне"""
    loop = asyncio.get_event_loop()
    return await loop.run_in_executor(None, train_model_sync, days_back)

# Эндпоинты
@app.post("/api/train")
async def train(request: TrainRequest, background_tasks: BackgroundTasks):
    """Запуск обучения модели"""
    global is_training
    
    if is_training:
        raise HTTPException(400, "Обучение уже выполняется")
    
    # Запускаем в фоне
    background_tasks.add_task(train_background, request.days_back)
    
    return {
        "message": "Обучение запущено",
        "days_back": request.days_back,
        "status": "training"
    }

@app.get("/api/predict")
async def predict(number: str):
    """Предсказание цены номера"""
    global predictor
    
    if predictor is None or predictor.model is None:
        raise HTTPException(503, "Модель не загружена")
    
    try:
        result = predictor.predict_single(number_str=number, return_features=True)
        
        if result is None:
            raise HTTPException(400, "Некорректный номер")

        return {
            "number": result['number'],
            "predicted_price": result['predicted_price'],
            "confidence": result['confidence'],
            "price_range": result['price_range'],
            "all_features": result['all_features']
        }
        
    except Exception as e:
        raise HTTPException(500, f"Ошибка: {str(e)}")

@app.get("/")
async def root():
    """Информация о сервисе"""
    return {
        "service": "Car Number Price API",
        "endpoints": {
            "POST /api/train": "Обучение модели",
            "GET /predict?number=": "Предсказание цены"
        }
    }

# Запуск сервера
if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)