import pandas as pd
import numpy as np
import psycopg2
from datetime import datetime, timedelta
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import LabelEncoder, StandardScaler
import joblib
import warnings
warnings.filterwarnings('ignore')

class DataLoader:
    def __init__(self, db_config):
        self.db_config = db_config
        self.conn = None

    def connect(self):
        """Подключение к PostgreSQL"""
        self.conn = psycopg2.connect(
            host=self.db_config['host'],
            database=self.db_config['database'],
            user=self.db_config['user'],
            password=self.db_config['password'],
            port=self.db_config['port']
        )

    def load_data(self, limit=None, days_back=365):
        """Загрузка данных из базы"""
        self.connect()

        # Загружаем данные за последний год (или все)
        query = f"""
        SELECT
            number,
            price,
            posted_at
        FROM car_numbers
        WHERE posted_at >= NOW() - INTERVAL '{days_back} days'
        AND price > 1000 AND price < 10000000  -- фильтр выбросов
        {'LIMIT ' + str(limit) if limit else ''}
        """

        df = pd.read_sql_query(query, self.conn)
        self.conn.close()

        print(f"Загружено {len(df)} записей")
        print(f"Диапазон цен: {df['price'].min():,.0f} - {df['price'].max():,.0f} руб.")
        print(f"Средняя цена: {df['price'].mean():,.0f} руб.")

        return df

    def close(self):
        if self.conn:
            self.conn.close()

