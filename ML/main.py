import os
import time
from datetime import datetime
from dotenv import load_dotenv
from data_loader import DataLoader
from price_predictor import NumberPricePredictor
from feature_extractor import FeatureEngineer

load_dotenv()

def main_pipeline():
    """Полный пайплайн обучения модели"""

    # Конфигурация БД
    db_config = {
        'host': os.getenv('DATABASE_HOST', 'postgres'),
        'database': os.getenv('DATABASE_NAME', 'postgres'),
        'user': os.getenv('DATABASE_USER', 'postgres'),
        'password': os.getenv('DATABASE_PASSWORD', 'postgres1'),
        'port': os.getenv('DATABASE_PORT', '5432')
    }

    # 1. Загрузка данных
    print("Шаг 1: Загрузка данных из PostgreSQL...")
    loader = DataLoader(db_config)
    raw_data = loader.load_data(days_back=365)  # данные за последний год
    loader.close()

    # 2. Подготовка признаков
    print("\nШаг 2: Подготовка признаков...")
    feature_engineer = FeatureEngineer()
    processed_data = feature_engineer.prepare_dataframe(raw_data)

    # 3. Обучение модели
    print("\nШаг 3: Обучение модели...")
    predictor = NumberPricePredictor()
    model = predictor.train(processed_data)

    # 4. Тестирование на примерах
    print("\nШаг 4: Тестирование модели...")
    test_numbers = [
        "Х500ОХ761",
    ]

    for number in test_numbers:
        result = predictor.predict_single(number)
        if result:
            print(f"\n{result['number']}:")
            print(f"  Предсказанная цена: {result['predicted_price']:,} руб.")
            print(f"  Диапазон: {result['price_range']['low']:,} - {result['price_range']['high']:,} руб.")
            print(f"  Уверенность: {result['confidence']:.0%}")

    return predictor

def get_price(number):
    predictor = NumberPricePredictor()
    obj = predictor.predict_single(number, True)
    print(obj)
    FeatureEngineer().analyze_number(number)


if __name__ == "__main__":
    main_pipeline()
    #predictor = get_price('С742РУ974')
