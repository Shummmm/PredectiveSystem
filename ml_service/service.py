from flask import Flask, request, jsonify
from flask_cors import CORS
import numpy as np
import joblib
import os

app = Flask(__name__)
CORS(app)  # Допуск запросов с любых доменов

# Прописываю пути к файлам модели
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
MODEL_PATH = os.path.join(BASE_DIR, 'model.pkl')
SCALER_PATH = os.path.join(BASE_DIR, 'scaler.pkl')

# Загрузка модели и scaler
print(" Загрузка ML модели...")
try:
    model = joblib.load(MODEL_PATH)
    scaler = joblib.load(SCALER_PATH)
    print("Модель и счетчик загружены успешно")
except Exception as e:
    print(f" Ошибка загрузки модели: {e}")
    print("   Запустите train_model.py сначала")
    model = None
    scaler = None

# Признаки 
FEATURE_NAMES = ['current', 'voltage', 'pressure', 'temperature', 'vibration']

@app.route('/health', methods=['GET'])
def health():
    """Проверка состояния сервиса"""
    return jsonify({
        'status': 'ok',
        'model_loaded': model is not None,
        'scaler_loaded': scaler is not None
    })

@app.route('/predict', methods=['POST'])
def predict():
    """Прогнозирование RUL по признакам"""
    if model is None or scaler is None:
        return jsonify({'error': 'Модель не загружена'}), 503
    
    data = request.get_json()
    
    if not data:
        return jsonify({'error': 'Нет данных'}), 400
    
    features = data.get('features', [])
    
    if not features:
        return jsonify({'error': 'Не переданы признаки (features)'}), 400
    
    # Проверка формата
    if isinstance(features[0], list):
        # Множество наблюдений: [[feat1, feat2, ...], [...]]
        X = np.array(features)
    else:
        # Одно наблюдение: [feat1, feat2, ...]
        X = np.array([features])
    
    # Проверка размерности
    if X.shape[1] != len(FEATURE_NAMES):
        return jsonify({
            'error': f'Неверное число признаков. Ожидается {len(FEATURE_NAMES)}, получено {X.shape[1]}'
        }), 400
    
    # Масштабирование + предсказание
    try:
        X_scaled = scaler.transform(X)
        predictions = model.predict(X_scaled)
        
        return jsonify({
            'predictions': predictions.tolist(),
            'count': len(predictions)
        })
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/predict_last', methods=['POST'])
def predict_last():
    """Прогнозирование RUL по последнему наблюдению (упрощенный вариант)"""
    if model is None or scaler is None:
        return jsonify({'error': 'Модель не загружена'}), 503
    
    data = request.get_json()
    
    if not data:
        return jsonify({'error': 'Нет данных'}), 400
    
    # Допуск двух форматов
    if 'features' in data:
        features = data['features']
    elif all(k in data for k in FEATURE_NAMES):
        features = [data[k] for k in FEATURE_NAMES]
    else:
        return jsonify({'error': f'Необходимы признаки: {FEATURE_NAMES}'}), 400
    
    if isinstance(features, list) and not isinstance(features[0], list):
        features = [features]
    
    X = np.array(features)
    
    if X.shape[1] != len(FEATURE_NAMES):
        return jsonify({'error': f'Неверное число признаков'}), 400
    
    try:
        X_scaled = scaler.transform(X)
        predictions = model.predict(X_scaled)
        
        return jsonify({
            'rul': int(predictions[0]),
            'predictions': predictions.tolist()
        })
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/feature_importance', methods=['GET'])
def feature_importance():
    """Возвращает важность признаков"""
    if model is None:
        return jsonify({'error': 'Модель не загружена'}), 503
    
    importance = model.feature_importances_
    return jsonify({
        feature: float(imp) 
        for feature, imp in zip(FEATURE_NAMES, importance)
    })

@app.route('/info', methods=['GET'])
def info():
    """Информация о сервисе и модели"""
    return jsonify({
        'service': 'ML Prediction Service',
        'version': '1.0',
        'model_loaded': model is not None,
        'features': FEATURE_NAMES,
        'model_type': type(model).__name__ if model else None
    })

if __name__ == '__main__':
    print("\n" + "="*50)
    print(" ML СЕРВИС ЗАПУЩЕН")
    print(f"   Порт: 5001")
    print(f"   Модель: {MODEL_PATH}")
    print(f"   Scaler: {SCALER_PATH}")
    print("="*50 + "\n")
    
    app.run(host='127.0.0.1', port=5001, debug=False)