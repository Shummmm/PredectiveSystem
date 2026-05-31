import pandas as pd
import numpy as np
from sklearn.ensemble import RandomForestRegressor
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import train_test_split, cross_val_score, KFold
from sklearn.metrics import mean_absolute_error, mean_squared_error, r2_score
from sklearn.pipeline import Pipeline
import joblib

# Генерация датасета из реальных данных
np.random.seed(42)
n_cycles = 3000
failure_cycle = 2500

print(" Генерация датасета...")
data = []

for cycle in range(n_cycles):
    current = 15000 + np.random.normal(0, 100)
    voltage = 2.5 + np.random.normal(0, 0.05)
    pressure = 6.0 + np.random.normal(0, 0.1)
    temperature = 30 + np.random.normal(0, 0.5)
    vibration = 1.0 + np.random.normal(0, 0.05)
    
    wear = cycle / failure_cycle
    current = current - 1500 * wear
    pressure = pressure - 2.0 * wear
    temperature = temperature + 20 * wear
    vibration = vibration + 1.5 * wear
    
    RUL = max(0, failure_cycle - cycle)
    data.append([current, voltage, pressure, temperature, vibration, RUL])

df = pd.DataFrame(data, columns=['current', 'voltage', 'pressure', 'temperature', 'vibration', 'RUL'])

# Признаки и цель
features = ['current', 'voltage', 'pressure', 'temperature', 'vibration']
X = df[features]
y = df['RUL']

# Разделение
X_train, X_test, y_train, y_test = train_test_split(
    X, y, test_size=0.2, random_state=42
)

# Pipeline 
pipeline = Pipeline([
    ('scaler', StandardScaler()),
    ('model', RandomForestRegressor(
        n_estimators=200,
        max_depth=15,
        random_state=42,
        n_jobs=-1
    ))
])

#  КРОСС-ВАЛИДАЦИЯ
print(" Кросс-валидация (5 фолдов)...")

kf = KFold(n_splits=5, shuffle=True, random_state=42)

mae_scores = cross_val_score(
    pipeline,
    X_train,
    y_train,
    cv=kf,
    scoring='neg_mean_absolute_error',
    n_jobs=-1
)

rmse_scores = cross_val_score(
    pipeline,
    X_train,
    y_train,
    cv=kf,
    scoring='neg_mean_squared_error',
    n_jobs=-1
)

# Привожу к нормальному виду
mae_scores = -mae_scores
rmse_scores = np.sqrt(-rmse_scores)

print("\n" + "="*50)
print("РЕЗУЛЬТАТЫ КРОСС-ВАЛИДАЦИИ")
print("="*50)
print(f"MAE:  {mae_scores.mean():.2f} ± {mae_scores.std():.2f}")
print(f"RMSE: {rmse_scores.mean():.2f} ± {rmse_scores.std():.2f}")
print("="*50)


# ОБУЧЕНИЕ НА ВСЕМ TRAIN
print(" Обучение финальной модели...")
pipeline.fit(X_train, y_train)

# Предсказания
y_pred = pipeline.predict(X_test)

#  МЕТРИКИ
mae = mean_absolute_error(y_test, y_pred)
rmse = np.sqrt(mean_squared_error(y_test, y_pred))

mask = y_test > 0
mape = np.mean(np.abs((y_test[mask] - y_pred[mask]) / y_test[mask])) * 100

r2 = r2_score(y_test, y_pred)

print("\n" + "="*50)
print("РЕЗУЛЬТАТЫ НА TEST")
print("="*50)
print(f"{'Метрика':<30} {'Значение':<15}")
print("-"*50)
print(f"{'MAE':<30} {mae:>10.2f} циклов")
print(f"{'RMSE':<30} {rmse:>10.2f} циклов")
print(f"{'MAPE':<30} {mape:>10.2f} %")
print(f"{'R²':<30} {r2:>10.4f}")
print("="*50)


joblib.dump(pipeline, 'model_pipeline.pkl')
print(" Pipeline (model + scaler) сохранён как model_pipeline.pkl")