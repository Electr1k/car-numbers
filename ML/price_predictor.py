from sklearn.calibration import LabelEncoder
from sklearn.model_selection import train_test_split
import CarNumberFeatureExtractor
from feature_extractor import FeatureEngineer
import numpy as np
import pandas as pd
import joblib

from catboost import CatBoostRegressor, Pool
from sklearn.metrics import mean_absolute_error, mean_absolute_percentage_error
import matplotlib.pyplot as plt
import seaborn as sns

class NumberPricePredictor:
    def __init__(self, model_path='models/price_catboost_model.cbm'):
        self.model_path = model_path
        self.model = None
        self.scaler = None
        self.label_encoders = {}
        self.feature_engineer = FeatureEngineer()
        
    def prepare_features(self, features_df):
        """Подготовка признаков для обучения"""
        
        # Копируем данные
        df = features_df.copy()
        
        # Удаляем целевую переменную
        if 'log_price' in df.columns:
            y = df['log_price'].values
            df = df.drop(['price', 'log_price'], axis=1)
        else:
            y = None
        
        # Определяем типы признаков
        categorical_features = ['digit_category', 'digit_type', 'region_group']
        numerical_features = [
            'digits', 'digit_1', 'digit_2', 'digit_3', 'digit_sum',
            'is_single_digit', 'is_triple', 'is_mirror', 'is_sequence',
            'is_round', 'has_7', 'has_0', 'has_0_middle',
            'letter1_num', 'letter2_num', 'letter3_num',
            'letter_diff_1_2', 'letter_diff_2_3',
            'is_triple_letters', 'is_vip_series', 'is_mirror_series',
            'is_beautiful_word', 'is_prestige_first_letter',
            'is_same_last_two_letters', 'is_hot_series',
            'region', 'is_moscow', 'is_spb', 'is_million_city', 'is_early_region',
            'region_length', 'region_last_two', 'region_last_digit', 'region_first_digit',
            'digit_region_exact_match', 'digit_region_last_two_match', 'digit_region_first_two_match',
            'visual_match_score', 'semantic_match', 'full_pattern_match',
            'golden_number', 'digit_letter_position_match',
            'prestige_score_raw', 'prestige_score'
        ]
        
        # Убираем признаки, которых может не быть
        available_numerical = [col for col in numerical_features if col in df.columns]
        available_categorical = [col for col in categorical_features if col in df.columns]
        
        # Кодируем категориальные признаки для CatBoost
        for col in available_categorical:
            le = LabelEncoder()
            df[col] = le.fit_transform(df[col].astype(str))
            self.label_encoders[col] = le

        # Отслеживаем, какие признаки использовались
        self.used_features = available_numerical + available_categorical
        print(f"Используется {len(self.used_features)} признаков:")
        print(f"  - Категориальные: {len(available_categorical)}")
        print(f"  - Числовые: {len(available_numerical)}")
        
        # Оставляем только нужные признаки
        df = df[self.used_features]
        
        # Создаем CatBoost Pool
        if y is not None:
            pool = Pool(df, y, cat_features=available_categorical)
        else:
            pool = Pool(df, cat_features=available_categorical)
        
        return df, y, pool

    def train(self, df, test_size=0.2, random_state=42):
        """Обучение модели"""
        
        print("=" * 50)
        print("НАЧАЛО ОБУЧЕНИЯ МОДЕЛИ")
        print("=" * 50)
        
        if 'log_price' not in df.columns and 'price' in df.columns:
            df['log_price'] = np.log1p(df['price'])

        # Разделение данных
        train_df, test_df = train_test_split(
            df, test_size=test_size, random_state=random_state, shuffle=True
        )
        
        print(f"Обучающая выборка: {len(train_df)} записей")
        print(f"Тестовая выборка: {len(test_df)} записей")
        
        # Подготовка данных
        X_train, y_train, train_pool = self.prepare_features(train_df)
        X_test, y_test, test_pool = self.prepare_features(test_df)
        
        # Параметры CatBoost
        model_params = {
            'iterations': 2000,
            'learning_rate': 0.03,
            'depth': 6,
            'l2_leaf_reg': 3,
            'loss_function': 'RMSE',
            'eval_metric': 'MAE',
            'early_stopping_rounds': 100,
            'random_seed': random_state,
            'verbose': 100,
            'task_type': 'CPU',  # 'GPU' если есть видеокарта
            'border_count': 128,
            'cat_features': [col for col in X_train.columns 
                           if col in ['digit_category', 'digit_type', 'region_group', 'prestige_category']]
        }
        
        # Создание и обучение модели
        self.model = CatBoostRegressor(**model_params)
        
        self.model.fit(
            train_pool,
            eval_set=test_pool,
            plot=False  # можно установить True для визуализации
        )
        
        # Оценка модели
        self.evaluate(X_test, y_test)
        
        # Сохранение модели
        self.save_model()
        
        return self.model
    
    def evaluate(self, X_test, y_test):
        """Оценка качества модели"""
        
        predictions_log = self.model.predict(X_test)
        predictions = np.expm1(predictions_log)
        actual = np.expm1(y_test)
        
        # Метрики
        mae = mean_absolute_error(actual, predictions)
        mape = mean_absolute_percentage_error(actual, predictions) * 100
        rmse = np.sqrt(np.mean((actual - predictions) ** 2))
        
        print("\n" + "=" * 50)
        print("РЕЗУЛЬТАТЫ ОЦЕНКИ МОДЕЛИ")
        print("=" * 50)
        print(f"MAE: {mae:,.0f} руб.")
        print(f"MAPE: {mape:.1f}%")
        print(f"RMSE: {rmse:,.0f} руб.")
        print(f"Средняя цена в тесте: {actual.mean():,.0f} руб.")
        print(f"Медианная цена в тесте: {np.median(actual):,.0f} руб.")
    
    
    def save_model(self):
        """Сохранение модели и препроцессоров"""
        import os
        os.makedirs('models', exist_ok=True)
        
        # Сохраняем CatBoost модель
        self.model.save_model(self.model_path)
        
        # Сохраняем label encoders
        joblib.dump(self.label_encoders, 'models/price_label_encoders.pkl')
        
        # Сохраняем feature engineer
        joblib.dump(self.feature_engineer, 'models/price_feature_engineer.pkl')
        
        if hasattr(self, 'used_features'):
            joblib.dump(self.used_features, 'models/used_features.pkl')
        else:
            # Сохраняем пустой список как fallback
            joblib.dump([], 'models/used_features.pkl')
        print(f"\nМодель сохранена в {self.model_path}")
    
    def load_model(self):
        """Загрузка модели"""
        self.model = CatBoostRegressor()
        self.model.load_model(self.model_path)
        self.label_encoders = joblib.load('models/price_label_encoders.pkl')
        self.feature_engineer = joblib.load('models/price_feature_engineer.pkl')

        try:
            self.used_features = joblib.load('models/used_features.pkl')
        except FileNotFoundError:
            print("Предупреждение: файл used_features.pkl не найден. Создаю пустой список признаков.")
            self.used_features = []
        print("Модель загружена")
    
    def predict_single(self, number_str, return_features=False):
        """Предсказание для одного номера"""
        if self.model is None:
            self.load_model()
        
        # Извлекаем признаки
        features = self.feature_engineer.extract_features(number_str)
        if features is None:
            return None
        
        # Преобразуем в DataFrame
        features_df = pd.DataFrame([features])
        
        # Подготовка признаков для предсказания
        df_processed = features_df.copy()
        
        # Кодируем категориальные признаки
        for col in self.label_encoders:
            if col in df_processed.columns:
                le = self.label_encoders[col]
                try:
                    df_processed[col] = le.transform([features[col]])[0]
                except ValueError:
                    # Если новое значение, используем most_frequent класс
                    df_processed[col] = le.transform([le.classes_[0]])[0]
        
        # Оставляем только признаки, использованные при обучении
        available_features = [col for col in self.used_features if col in df_processed.columns]
        missing_features = set(self.used_features) - set(available_features)
        
        if missing_features:
            print(f"Предупреждение: отсутствуют признаки: {missing_features}")
            for feature in missing_features:
                df_processed[feature] = 0
        
        df_processed = df_processed[self.used_features]

        
        # Предсказание
        prediction_log = self.model.predict(df_processed)[0]
        prediction = np.expm1(prediction_log)
        
        # Оценка уверенности
        confidence = self._estimate_confidence(features)
        
        result = {
            'number': number_str,
            'predicted_price': int(round(prediction, -2)),  # округляем до сотен
            'confidence': confidence,
            'price_range': {
                'low': int(round(prediction * 0.8, -2)),
                'high': int(round(prediction * 1.2, -2))
            },
            'features_summary': {
                'digit_type': features.get('digit_category', 'unknown'),
                'full_series': features.get('full_series', ''),
                'region': features.get('region', 0),
                'prestige_score': features.get('prestige_score', 0),
                'prestige_category': features.get('prestige_category', 'unknown'),
                'is_vip': bool(features.get('is_vip_series', 0)),
                'is_moscow': bool(features.get('is_moscow', 0)),
                'golden_number': bool(features.get('golden_number', 0))
            }
        }
        
        if return_features:
            result['all_features'] = features
            
        return result
        
        
    def _estimate_confidence(self, features):
        """Оценка уверенности предсказания на основе признаков"""
        confidence = 0.7  # базовая уверенность
        
        # Увеличиваем уверенность для "хорошо описанных" номеров
        if features.get('digit_category') in ['premium', 'prestige']:
            confidence += 0.1
        if features.get('is_vip_series'):
            confidence += 0.05
        if features.get('is_moscow'):
            confidence += 0.05
        if features.get('golden_number'):
            confidence += 0.1
        if features.get('prestige_score', 0) > 70:
            confidence += 0.1
            
        return min(confidence, 0.95)