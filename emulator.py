import sqlite3
import random
import time
from datetime import datetime, timedelta
import os

DB_PATH = 'sensor_data.db'

def init_db():
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS sensor_readings (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            timestamp TEXT NOT NULL,
            current REAL,
            voltage REAL,
            pressure REAL,
            temperature REAL,
            vibration REAL,
            is_processed INTEGER DEFAULT 0
        )
    ''')
    conn.commit()
    conn.close()

def generate_reading(offset_seconds=0):
    """Генерирует показание датчиков с возможным смещением времени"""
    current = 15000 + random.gauss(0, 50)
    voltage = 2.5 + random.gauss(0, 0.03)
    pressure = 6.0 + random.gauss(0, 0.05)
    temperature = 35 + random.gauss(0, 1)
    vibration = 1.0 + random.gauss(0, 0.05)
    
    # Время с учётом смещения
    dt = datetime.now() + timedelta(seconds=offset_seconds)
    
    return {
        'timestamp': dt.strftime("%Y-%m-%d %H:%M:%S"),
        'current': round(current, 2),
        'voltage': round(voltage, 2),
        'pressure': round(pressure, 2),
        'temperature': round(temperature, 1),
        'vibration': round(vibration, 3)
    }

def add_reading(reading):
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    cursor.execute('''
        INSERT INTO sensor_readings 
        (timestamp, current, voltage, pressure, temperature, vibration, is_processed)
        VALUES (?, ?, ?, ?, ?, ?, 0)
    ''', (
        reading['timestamp'],
        reading['current'],
        reading['voltage'],
        reading['pressure'],
        reading['temperature'],
        reading['vibration']
    ))
    conn.commit()
    conn.close()

def main():
    init_db()
    print(" ЭМУЛЯТОР ДАТЧИКОВ ЗАПУЩЕН")
    print("   Генерирует данные каждые 10 секунд")
    print("   Ctrl+C для остановки\n")
    
    try:
        count = 0
        while True:
            reading = generate_reading()
            add_reading(reading)
            count += 1
            print(f"[{count:4d}] {reading['timestamp']} | "
                  f"Ток: {reading['current']:6.0f}A | "
                  f"P: {reading['pressure']:.2f}МПа | "
                  f"T: {reading['temperature']:.1f}°C")
            time.sleep(10)
    except KeyboardInterrupt:
        print("\n Эмулятор остановлен")
        print(f"   Всего добавлено записей: {count}")

if __name__ == "__main__":
    main()
