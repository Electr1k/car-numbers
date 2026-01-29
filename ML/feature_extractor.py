import pandas as pd
import numpy as np
from typing import Dict, Any, Optional, List, Tuple
import re

class FeatureEngineer:
    def __init__(self):
        # Все возможные буквы в российских номерах (12 букв)
        self.allowed_letters = set('АВЕКМНОРСТУХ')
        self.letter_to_num = {letter: idx+1 for idx, letter in enumerate(sorted(self.allowed_letters))}
        
        # --- БУКВЕННАЯ ЧАСТЬ ---
        # Полные серии (3 буквы) - категории
        self.prestige_series_3 = {
            # Одинаковые буквы (топ)
            'ААА', 'МММ', 'ООО', 'ССС', 'ХХХ', 'ТТТ', 'ВВВ', 'РРР',
            'ЕЕЕ', 'ККК', 'ННН', 'УУУ',
        }
        
        # особые серии
        self.vip_series = {
            'АМР', 'ЕКХ', 'ВОР', 'АУЕ', 'ХКХ', 'РМР', 'МРМ',
            'САС', 'ТТР', 'УКХ', 'ХАМ', 'ММР', 'ОМР'
        }
        
        # Красивые словесные комбинации
        self.beautiful_words = {
            'АМР', 'ТОР', 'СТО', 'УМР', 'ХОР', 'РАМ', 'МАР', 'РОТ',
            'РАК', 'КАТ', 'ТАК', 'МАК', 'РАН', 'НОР', 'ТОР', 'УРА',
            'МОР', 'РОК', 'КОТ', 'ТОК', 'МУХ', 'ХАМ', 'СОН', 'НОС',
            'РОС', 'СОР', 'ТУТ', 'ТАМ', 'ТОМ'
        }
        
        # Зеркальные серии
        self.mirror_series = {
            'АВА', 'АКА', 'АМА', 'АНА', 'АРА', 'АСА', 'АТА', 'АХА',
            'МАМ', 'МОМ', 'МУМ', 'ОКО', 'ОМО', 'ОРО', 'ОСО', 'ОТО',
            'САС', 'СОС', 'СУС', 'ТАТ', 'ТОТ', 'ТУТ', 'ХАХ', 'ХОХ'
        }
        
        # Буквы, особо ценные в первой позиции
        self.prestige_first_letters = {'А', 'М', 'О', 'С', 'Х', 'Т', 'В'}
        
        # --- ЦИФРОВАЯ ЧАСТЬ ---
        # Категории цифровых комбинаций
        self.digit_categories = {
            # Премиум (макс. вес)
            'single_digit': set(range(1, 10)),  # 001-009
            'triple': {111, 222, 333, 444, 555, 666, 777, 888, 999},
            
            # Престиж (высокий вес)
            'mirror': {101, 111, 121, 131, 141, 151, 161, 171, 181, 191,
                      202, 212, 222, 232, 242, 252, 262, 272, 282, 292,
                      303, 313, 323, 333, 343, 353, 363, 373, 383, 393,
                      404, 414, 424, 434, 444, 454, 464, 474, 484, 494,
                      505, 515, 525, 535, 545, 555, 565, 575, 585, 595,
                      606, 616, 626, 636, 646, 656, 666, 676, 686, 696,
                      707, 717, 727, 737, 747, 757, 767, 777, 787, 797,
                      808, 818, 828, 838, 848, 858, 868, 878, 888, 898,
                      909, 919, 929, 939, 949, 959, 969, 979, 989, 999},
            'sequence': {123, 234, 345, 456, 567, 678, 789,
                        987, 876, 765, 654, 543, 432, 321},
            'round': {100, 200, 300, 400, 500, 600, 700, 800, 900},
            
            # Популярные (средний вес)
            'lucky': {7, 77, 777, 13, 88, 888, 99, 999},
            'popular': {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 17, 18, 19,
                       22, 33, 44, 55, 66, 77, 88, 99, 100, 111, 123, 200,
                       222, 234, 300, 333, 345, 400, 444, 456, 500, 555,
                       567, 600, 666, 678, 700, 777, 789, 800, 888, 900, 999}
        }
        
        # --- РЕГИОНЫ ---
        # Москва
        self.moscow_codes = {
            # Основные
            77, 97, 99, 177, 197, 199, 777, 797, 799,
            # Дополнительные
            277, 297, 299, 377, 397, 399, 477, 497, 499,
            577, 597, 599, 677, 697, 699, 877, 897, 899,
            977, 997, 999
        }
        
        # Санкт-Петербург
        self.spb_codes = {78, 98, 178, 198}
        
        # Города-миллионники с актуальными кодами
        self.million_cities = {
            # Краснодарский край
            23, 93, 123,
            # Новосибирск
            54, 154,
            # Пермь
            59, 81, 159,
            # Ростов-на-Дону
            61, 161, 761,
            # Самара
            63, 163, 763,
            # Саратов
            64, 164,
            # Екатеринбург
            66, 96, 196,
            # Тюмень
            72,
            # Челябинск
            74, 174,
            # Уфа
            2, 102, 702,
            # Казань
            16, 116, 716,
            # Нижний Новгород
            52, 152,
            # Волгоград
            34, 134
        }
        
        # Ранние/престижные регионы (01-39)
        self.early_regions = set(range(1, 40))
        
        # --- ВЗАИМОДЕЙСТВИЯ ---
        # Визуальные аналогии
        self.visual_analogies = {
            '0': ['О'],  # 0 похож на О
            '1': ['Т'],  # 1 похож на Т (вертикальная линия)
            '7': ['У'],  # 7 похож на У (перевернутый)
            '8': ['В'],  # 8 похож на В
        }
        
        # Семантические совпадения (цифры + буквы)
        self.semantic_matches = {
        }
        
        # --- ВЕСА ДЛЯ РАСЧЕТА ПРЕСТИЖНОСТИ ---
        self.weights = {
            # Цифры (40%)
            'single_digit': 25,
            'triple_digits': 20,
            'mirror_digits': 15,
            'sequence_digits': 12,
            'round_number': 10,
            'lucky_number': 8,
            'popular_number': 5,
            'has_7': 3,
            'has_0': 5,
            
            # Буквы (30%)
            'triple_letters': 20,
            'vip_series': 18,
            'mirror_series': 15,
            'beautiful_word': 12,
            'prestige_first_letter': 3,
            'same_last_two_letters': 5,
            
            # Регион (20%)
            'moscow_region': 20,
            'spb_region': 15,
            'million_city': 12,
            'early_region': 8,
            
            # Взаимодействия (10%)
            'full_pattern_match': 10,
            'semantic_match': 0,
            'visual_match': 6,
            'digit_region_exact_match': 0,
            'digit_region_partial_match': 0,
            'golden_number': 15,  # особо ценные комбинации
        }
        
        # Максимальные баллы для нормализации
        self.max_scores = {
            'digits': 25 + 20 + 15 + 12 + 10 + 8 + 5 + 3 + 5,  # 103
            'letters': 20 + 18 + 15 + 12 + 8 + 5,  # 78
            'region': 20 + 15 + 12 + 8,  # 55
            'interactions': 10 + 0 + 6 + 0 + 0 + 15,  # 31
        }

    def extract_features(self, number_str: str) -> Optional[Dict[str, Any]]:
        """Основной метод извлечения признаков из номера"""
        features = {}
        
        # 1. РАЗБИЕНИЕ НОМЕРА
        if not self._validate_and_parse(number_str, features):
            return None
            
        # 2. ЦИФРОВАЯ ЧАСТЬ
        self._extract_digit_features(features)
        
        # 3. БУКВЕННАЯ ЧАСТЬ
        self._extract_letter_features(features)
        
        # 4. РЕГИОНЫ
        self._extract_region_features(features)
        
        # 5. ВЗАИМОДЕЙСТВИЯ
        self._extract_interaction_features(features)
        
        # 6. РАСЧЕТ ПРЕСТИЖНОСТИ
        self._calculate_prestige_score(features)
        
        return features
    
    def _validate_and_parse(self, number_str: str, features: Dict[str, Any]) -> bool:
        """Проверка формата и разбор номера"""
        pattern = r'^[АВЕКМНОРСТУХ]\d{3}[АВЕКМНОРСТУХ]{2}\d{2,3}$'
        if not re.match(pattern, number_str):
            return False
        
        try:
            features['original_number'] = number_str
            features['first_letter'] = number_str[0]
            features['digits_str'] = number_str[1:4]
            features['series'] = number_str[4:6]  # последние 2 буквы
            features['region_str'] = number_str[6:]
            
            features['digits'] = int(features['digits_str'])
            features['region'] = int(features['region_str'])
            features['full_series'] = features['first_letter'] + features['series']
            
            # Отдельные цифры
            d1, d2, d3 = features['digits_str']
            features['digit_1'] = int(d1)
            features['digit_2'] = int(d2)
            features['digit_3'] = int(d3)
            
            return True
        except Exception as e:
            print(f"Ошибка разбора {number_str}: {e}")
            return False
    
    def _extract_digit_features(self, features: Dict[str, Any]) -> None:
        """Извлечение признаков из цифровой части"""
        d = features['digits']
        d1, d2, d3 = features['digit_1'], features['digit_2'], features['digit_3']
        
        # Бинарные признаки
        features['is_single_digit'] = 1 if d <= 9 else 0
        features['is_triple'] = 1 if d1 == d2 == d3 else 0
        features['is_mirror'] = 1 if d1 == d3 else 0
        features['is_sequence'] = 1 if (abs(d2 - d1) == 1 and 
                                       abs(d3 - d2) == 1 and 
                                       (d2 > d1) == (d3 > d2)) else 0
        features['is_round'] = 1 if d % 100 == 0 else 0
        features['has_7'] = 1 if '7' in features['digits_str'] else 0
        features['has_0'] = 1 if '0' in features['digits_str'] else 0
        features['has_0_middle'] = 1 if d2 == 0 else 0  # 0 в середине
        
        # Категория цифр
        digit_type = 'regular'
        for category, values in self.digit_categories.items():
            if d in values:
                digit_type = category
                break
        
        features['digit_type'] = digit_type
        
        # Группа категорий для упрощения
        if digit_type in ['single_digit', 'triple']:
            features['digit_category'] = 'premium'
        elif digit_type in ['mirror', 'sequence', 'round', 'lucky']:
            features['digit_category'] = 'prestige'
        elif digit_type == 'popular':
            features['digit_category'] = 'popular'
        else:
            features['digit_category'] = 'standard'
    
    def _extract_letter_features(self, features: Dict[str, Any]) -> None:
        """Извлечение признаков из буквенной части"""
        full_series = features['full_series']
        first_letter = features['first_letter']
        series = features['series']
        
        # Базовые признаки
        features['letter1_num'] = self.letter_to_num.get(full_series[0], 0)
        features['letter2_num'] = self.letter_to_num.get(full_series[1], 0)
        features['letter3_num'] = self.letter_to_num.get(full_series[2], 0)
        
        # Категории буквенных комбинаций
        features['is_triple_letters'] = 1 if (full_series[0] == full_series[1] == full_series[2]) else 0
        features['is_vip_series'] = 1 if full_series in self.vip_series else 0
        features['is_mirror_series'] = 1 if full_series in self.mirror_series else 0
        features['is_beautiful_word'] = 1 if full_series in self.beautiful_words else 0
        features['is_prestige_first_letter'] = 1 if first_letter in self.prestige_first_letters else 0
        features['is_same_last_two_letters'] = 1 if series[0] == series[1] else 0
        
        # Дополнительные буквенные признаки
        features['letter_diff_1_2'] = abs(features['letter1_num'] - features['letter2_num'])
        features['letter_diff_2_3'] = abs(features['letter2_num'] - features['letter3_num'])
        
        # Является ли серия "горячей" (популярные комбинации)
        features['is_hot_series'] = 1 if (features['is_vip_series'] or 
                                         features['is_beautiful_word'] or 
                                         features['is_mirror_series']) else 0
    
    def _extract_region_features(self, features: Dict[str, Any]) -> None:
        """Извлечение признаков региона"""
        region = features['region']
        
        # Бинарные признаки региона
        features['is_moscow'] = 1 if region in self.moscow_codes else 0
        features['is_spb'] = 1 if region in self.spb_codes else 0
        features['is_million_city'] = 1 if region in self.million_cities else 0
        features['is_early_region'] = 1 if region in self.early_regions else 0
        
        # Группа региона
        if features['is_moscow']:
            features['region_group'] = 'moscow'
        elif features['is_spb']:
            features['region_group'] = 'spb'
        elif features['is_million_city']:
            features['region_group'] = 'million'
        elif features['is_early_region']:
            features['region_group'] = 'early'
        else:
            features['region_group'] = 'other'
        
        # Дополнительные признаки региона
        features['region_length'] = len(features['region_str'])
        features['region_last_two'] = region % 100
        features['region_last_digit'] = region % 10
        features['region_first_digit'] = region // 100 if region >= 100 else region // 10
    
    def _extract_interaction_features(self, features: Dict[str, Any]) -> None:
        """Извлечение признаков взаимодействий между частями номера"""
        digits_str = features['digits_str']
        full_series = features['full_series']
        region = features['region']
        digits = features['digits']
        
        # 1. Совпадение цифр и региона
        features['digit_region_exact_match'] = 1 if digits == region else 0
        features['digit_region_last_two_match'] = 0
        features['digit_region_first_two_match'] = 0
        
        # 2. Визуальные аналогии (цифры-буквы)
        visual_matches = 0
        for digit, letters in self.visual_analogies.items():
            if digit in digits_str:
                for letter in letters:
                    if letter in full_series:
                        visual_matches += 1
        features['visual_match_score'] = visual_matches
        
        # 3. Семантические совпадения
        semantic_match = 0
        for digit_pattern, letter_patterns in self.semantic_matches.items():
            if digit_pattern in digits_str:
                for letter_pattern in letter_patterns:
                    if letter_pattern in full_series:
                        semantic_match = 1
                        break
        features['semantic_match'] = semantic_match
        
        # 4. Полные паттерны (зеркальные цифры + зеркальные буквы)
        features['full_pattern_match'] = 0
        if features['is_mirror'] and features['is_mirror_series']:
            features['full_pattern_match'] = 1
        elif features['is_triple'] and features['is_triple_letters']:
            features['full_pattern_match'] = 2  # максимальный балл
        
        # 5. "Золотые" номера (особо ценные комбинации)
        features['golden_number'] = 0
        # Правило 1: Премиальные цифры + VIP серия + Москва
        if (features['digit_category'] == 'premium' and 
            features['is_vip_series'] and 
            features['is_moscow']):
            features['golden_number'] = 1
        # Правило 2: Три одинаковые цифры + три одинаковые буквы
        elif features['is_triple'] and features['is_triple_letters']:
            features['golden_number'] = 1
        
        # 6. Совпадение цифр и позиций букв
        features['digit_letter_position_match'] = 0
        # Проверяем, совпадает ли цифра с номером буквы в алфавите
        for i, digit_char in enumerate(digits_str):
            if i < 3:  # только для 3 букв
                letter_num = self.letter_to_num.get(full_series[i], 0)
                if letter_num == int(digit_char):
                    features['digit_letter_position_match'] += 1
    
    def _calculate_prestige_score(self, features: Dict[str, Any]) -> None:
        """Эвристический расчет престижности номера"""
        score = 0
        
        # --- ЦИФРЫ (40%) ---
        if features['is_single_digit']:
            score += self.weights['single_digit']
        if features['is_triple']:
            score += self.weights['triple_digits']
        if features['is_mirror']:
            score += self.weights['mirror_digits']
        if features['is_sequence']:
            score += self.weights['sequence_digits']
        if features['is_round']:
            score += self.weights['round_number']
        
        # Дополнительные цифровые признаки
        if features['digit_type'] == 'lucky':
            score += self.weights['lucky_number']
        elif features['digit_type'] == 'popular':
            score += self.weights['popular_number']
        
        if features['has_7']:
            score += self.weights['has_7']
        if features['has_0_middle']:
            score += self.weights['has_0']
        
        # --- БУКВЫ (30%) ---
        if features['is_triple_letters']:
            score += self.weights['triple_letters']
        if features['is_vip_series']:
            score += self.weights['vip_series']
        if features['is_mirror_series']:
            score += self.weights['mirror_series']
        if features['is_beautiful_word']:
            score += self.weights['beautiful_word']
        if features['is_prestige_first_letter']:
            score += self.weights['prestige_first_letter']
        if features['is_same_last_two_letters']:
            score += self.weights['same_last_two_letters']
        
        # --- РЕГИОН (20%) ---
        if features['is_moscow']:
            score += self.weights['moscow_region']
        elif features['is_spb']:
            score += self.weights['spb_region']
        
        if features['is_million_city']:
            score += self.weights['million_city']
        elif features['is_early_region']:
            score += self.weights['early_region']
        
        # --- ВЗАИМОДЕЙСТВИЯ (10%) ---
        if features['full_pattern_match'] > 0:
            score += self.weights['full_pattern_match'] * features['full_pattern_match']
        
        if features['semantic_match']:
            score += self.weights['semantic_match']
        
        if features['visual_match_score'] > 0:
            score += self.weights['visual_match'] * features['visual_match_score']
        
        if features['digit_region_exact_match']:
            score += self.weights['digit_region_exact_match']
        elif features['digit_region_last_two_match'] or features['digit_region_first_two_match']:
            score += self.weights['digit_region_partial_match']
        
        if features['golden_number'] > 0:
            score += self.weights['golden_number'] * features['golden_number']
        
        # Нормализация к 100 баллам
        features['prestige_score_raw'] = score
        features['prestige_score'] = min(int((float(score) / 250) * 100), 100)
        
        # Категория престижности
        if features['prestige_score'] >= 85:
            features['prestige_category'] = 'luxury'
        elif features['prestige_score'] >= 70:
            features['prestige_category'] = 'premium'
        elif features['prestige_score'] >= 50:
            features['prestige_category'] = 'prestige'
        elif features['prestige_score'] >= 30:
            features['prestige_category'] = 'standard'
        else:
            features['prestige_category'] = 'economy'
    
    def prepare_dataframe(self, df: pd.DataFrame, number_col: str = 'number', 
                         price_col: str = 'price') -> pd.DataFrame:
        """Подготовка DataFrame с признаками для обучения"""
        features_list = []
        
        for idx, row in df.iterrows():
            number = row[number_col]
            features = self.extract_features(number)
            
            if features:
                # Сохраняем исходную цену
                if price_col in row:
                    features['price'] = float(row[price_col])
                    features['log_price'] = np.log1p(features['price'])
                
                features_list.append(features)
        
        if not features_list:
            return pd.DataFrame()
        
        features_df = pd.DataFrame(features_list)
        
        # Добавляем one-hot кодирование для категориальных признаков
        categorical_cols = ['digit_category', 'region_group', 'prestige_category']
        for col in categorical_cols:
            if col in features_df.columns:
                dummies = pd.get_dummies(features_df[col], prefix=col)
                features_df = pd.concat([features_df, dummies], axis=1)
        
        print(f"Успешно обработано {len(features_df)} из {len(df)} номеров")
        return features_df
    
    def analyze_number(self, number_str: str) -> Optional[Dict[str, Any]]:
        """Детальный анализ номера с выводом информации"""
        features = self.extract_features(number_str)
        
        if not features:
            print(f"Ошибка: некорректный номер {number_str}")
            return None
        
        print(f"\n{'='*60}")
        print(f"📊 АНАЛИЗ НОМЕРА: {number_str}")
        print(f"{'='*60}")
        
        print(f"\n🔷 СТРУКТУРА:")
        print(f"   Буквы: {features['full_series'][0]} {features['full_series'][1]}{features['full_series'][2]}")
        print(f"   Цифры: {features['digits_str']}")
        print(f"   Регион: {features['region_str']}")
        
        print(f"\n🔷 ЦИФРЫ ({features['digits_str']}):")
        print(f"   Тип: {features['digit_type']}")
        print(f"   Категория: {features['digit_category']}")
        
        print(f"\n🔷 БУКВЫ ({features['full_series']}):")
        if features['is_triple_letters']:
            print(f"   Одинаковые буквы: ДА")
        if features['is_vip_series']:
            print(f"   VIP серия: ДА")
        if features['is_mirror_series']:
            print(f"   Зеркальная серия: ДА")
        if features['is_beautiful_word']:
            print(f"   Красивое слово: ДА")
        
        print(f"\n🔷 РЕГИОН ({features['region']}):")
        if features['is_moscow']:
            print(f"   Москва: ДА")
        elif features['is_spb']:
            print(f"   СПб: ДА")
        elif features['is_million_city']:
            print(f"   Город-миллионник: ДА")
        elif features['is_early_region']:
            print(f"   Ранний регион: ДА")
        
        print(f"\n🔷 ВЗАИМОДЕЙСТВИЯ:")
        if features['full_pattern_match']:
            print(f"   Полный паттерн: {features['full_pattern_match']} балл(ов)")
        if features['semantic_match']:
            print(f"   Семантическое совпадение: ДА")
        if features['visual_match_score']:
            print(f"   Визуальные аналогии: {features['visual_match_score']}")
        if features['golden_number']:
            print(f"   Золотой номер: ДА (уровень {features['golden_number']})")
        
        print(f"\n🔷 ИТОГОВАЯ ОЦЕНКА:")
        print(f"   Престижность: {features['prestige_score']}/100")
        print(f"   Категория: {features['prestige_category']}")
        
        print(f"\n{'='*60}")
        
        return features


# Пример использования
if __name__ == "__main__":
    feature_engineer = FeatureEngineer()
    
    # Тестовые номера
    test_numbers = [
        "Х500ОХ761",   # Престижный московский
    ]
    
    for number in test_numbers:
        features = feature_engineer.analyze_number(number)
        if features:
            print(f"\nКлючевые признаки для {number}:")
            print(f"  - Престижность: {features['prestige_score']}")
            print(f"  - Категория: {features['prestige_category']}")
            print(f"  - VIP серия: {'ДА' if features['is_vip_series'] else 'НЕТ'}")
            print(f"  - Москва: {'ДА' if features['is_moscow'] else 'НЕТ'}")