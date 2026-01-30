import os
import csv
import logging
from datetime import datetime
import psycopg2
from psycopg2.extras import execute_values

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Configuration
DATABASE_URL = os.getenv('DATABASE_URL', 'postgresql://postgres:postgres@postgres:5432/microservices')

def get_db_connection():
    """Create database connection"""
    try:
        conn = psycopg2.connect(DATABASE_URL)
        return conn
    except Exception as e:
        logger.error(f"Database connection failed: {e}")
        return None

def create_table_if_not_exists(conn):
    """Create housing_price_index table if it doesn't exist"""
    try:
        cursor = conn.cursor()
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS housing_price_index (
                id SERIAL PRIMARY KEY,
                tarih DATE NOT NULL,
                istanbul_turkiye VARCHAR(50) NOT NULL,
                yeni_yeni_olmayan_konut VARCHAR(50) NOT NULL,
                fiyat_endeksi DECIMAL(10, 2) NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                UNIQUE(tarih, istanbul_turkiye, yeni_yeni_olmayan_konut)
            );
            
            CREATE INDEX IF NOT EXISTS idx_housing_tarih ON housing_price_index(tarih);
            CREATE INDEX IF NOT EXISTS idx_housing_location ON housing_price_index(istanbul_turkiye);
        """)
        conn.commit()
        cursor.close()
        logger.info("Table housing_price_index created or already exists")
    except Exception as e:
        logger.error(f"Error creating table: {e}")
        conn.rollback()
        raise

def import_csv_data(csv_file_path):
    """Import CSV data into PostgreSQL with duplicate prevention"""
    conn = get_db_connection()
    if not conn:
        logger.error("Failed to connect to database")
        return False
    
    try:
        # Create table if not exists
        create_table_if_not_exists(conn)
        
        # Read CSV file
        with open(csv_file_path, 'r', encoding='utf-8') as csvfile:
            csv_reader = csv.DictReader(csvfile)
            
            # Prepare data for batch insert
            data_to_insert = []
            for row in csv_reader:
                data_to_insert.append((
                    row['tarih'],
                    row['istanbul_turkiye'],
                    row['yeni_yeni_olmayan_konut'],
                    float(row['fiyat_endeksi'])
                ))
            
            # Use UPSERT (INSERT ... ON CONFLICT) to prevent duplicates
            cursor = conn.cursor()
            
            insert_query = """
                INSERT INTO housing_price_index 
                (tarih, istanbul_turkiye, yeni_yeni_olmayan_konut, fiyat_endeksi)
                VALUES %s
                ON CONFLICT (tarih, istanbul_turkiye, yeni_yeni_olmayan_konut) 
                DO UPDATE SET 
                    fiyat_endeksi = EXCLUDED.fiyat_endeksi,
                    updated_at = CURRENT_TIMESTAMP
            """
            
            execute_values(cursor, insert_query, data_to_insert)
            conn.commit()
            
            logger.info(f"Successfully imported {len(data_to_insert)} rows")
            logger.info(f"Rows affected: {cursor.rowcount}")
            
            cursor.close()
            return True
            
    except Exception as e:
        logger.error(f"Error importing data: {e}")
        conn.rollback()
        return False
    finally:
        conn.close()

def import_csv_string(csv_string):
    """Import CSV data from string into PostgreSQL with duplicate prevention"""
    conn = get_db_connection()
    if not conn:
        logger.error("Failed to connect to database")
        return False
    
    try:
        # Create table if not exists
        create_table_if_not_exists(conn)
        
        # Parse CSV string
        csv_lines = csv_string.strip().split('\n')
        csv_reader = csv.DictReader(csv_lines)
        
        # Prepare data for batch insert
        data_to_insert = []
        for row in csv_reader:
            data_to_insert.append((
                row['tarih'],
                row['istanbul_turkiye'],
                row['yeni_yeni_olmayan_konut'],
                float(row['fiyat_endeksi'])
            ))
        
        # Use UPSERT (INSERT ... ON CONFLICT) to prevent duplicates
        cursor = conn.cursor()
        
        insert_query = """
            INSERT INTO housing_price_index 
            (tarih, istanbul_turkiye, yeni_yeni_olmayan_konut, fiyat_endeksi)
            VALUES %s
            ON CONFLICT (tarih, istanbul_turkiye, yeni_yeni_olmayan_konut) 
            DO UPDATE SET 
                fiyat_endeksi = EXCLUDED.fiyat_endeksi,
                updated_at = CURRENT_TIMESTAMP
        """
        
        execute_values(cursor, insert_query, data_to_insert)
        conn.commit()
        
        logger.info(f"Successfully imported {len(data_to_insert)} rows")
        logger.info(f"Rows affected: {cursor.rowcount}")
        
        cursor.close()
        return True
        
    except Exception as e:
        logger.error(f"Error importing data: {e}")
        conn.rollback()
        return False
    finally:
        conn.close()

if __name__ == '__main__':
    import sys
    
    if len(sys.argv) > 1:
        # Import from file
        csv_file = sys.argv[1]
        logger.info(f"Importing data from {csv_file}")
        success = import_csv_data(csv_file)
    else:
        # Example CSV data (can be replaced with actual file path)
        csv_data = """tarih,istanbul_turkiye,yeni_yeni_olmayan_konut,fiyat_endeksi
2010-01-01,İstanbul,Yeni Konut,35.9
2010-01-01,İstanbul,Yeni Olmayan Konut,35.9
2010-01-01,Türkiye,Yeni Konut,44.9
2010-01-01,Türkiye,Yeni Olmayan Konut,45.3
2010-02-01,İstanbul,Yeni Konut,36.6
2010-02-01,İstanbul,Yeni Olmayan Konut,36.1
2010-02-01,Türkiye,Yeni Konut,45.3
2010-02-01,Türkiye,Yeni Olmayan Konut,45.5
2010-03-01,İstanbul,Yeni Konut,37.4
2010-03-01,İstanbul,Yeni Olmayan Konut,36.3
2010-03-01,Türkiye,Yeni Konut,45.7
2010-03-01,Türkiye,Yeni Olmayan Konut,46.0
2010-04-01,İstanbul,Yeni Konut,38.0
2010-04-01,İstanbul,Yeni Olmayan Konut,36.7
2010-04-01,Türkiye,Yeni Konut,45.9
2010-04-01,Türkiye,Yeni Olmayan Konut,46.3
2010-05-01,İstanbul,Yeni Konut,38.0
2010-05-01,İstanbul,Yeni Olmayan Konut,37.1
2010-05-01,Türkiye,Yeni Konut,46.1
2010-05-01,Türkiye,Yeni Olmayan Konut,46.6
2010-06-01,İstanbul,Yeni Konut,37.6
2010-06-01,İstanbul,Yeni Olmayan Konut,37.5
2010-06-01,Türkiye,Yeni Konut,46.3
2010-06-01,Türkiye,Yeni Olmayan Konut,46.8"""
        
        logger.info("Importing example CSV data")
        success = import_csv_string(csv_data)
    
    if success:
        logger.info("Data import completed successfully")
    else:
        logger.error("Data import failed")
        sys.exit(1)
