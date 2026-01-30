import os
import json
import time
import logging
from datetime import datetime
from flask import Flask, request, jsonify
from flask_cors import CORS
import psycopg2
from psycopg2.extras import RealDictCursor
from psycopg2.extras import execute_values
import redis
from threading import Thread
import uuid
import csv
from io import StringIO

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

app = Flask(__name__)
CORS(app)

# Configuration
DATABASE_URL = os.getenv('DATABASE_URL', 'postgresql://postgres:postgres@postgres:5432/microservices')
REDIS_HOST = os.getenv('REDIS_HOST', 'redis')
REDIS_PORT = int(os.getenv('REDIS_PORT', '6379'))

# Initialize Redis
redis_client = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, db=0, decode_responses=True)

def get_db_connection():
    """Create database connection"""
    try:
        conn = psycopg2.connect(DATABASE_URL)
        return conn
    except Exception as e:
        logger.error(f"Database connection failed: {e}")
        return None

def process_data_worker(job_id, data):
    """Background worker to process data"""
    try:
        logger.info(f"Starting job {job_id}")
        
        # Update job status to processing
        redis_client.hset(f"job:{job_id}", mapping={
            "status": "processing",
            "started_at": datetime.now().isoformat()
        })
        
        # Simulate data processing (e.g., analytics, transformations)
        time.sleep(3)  # Simulate work
        
        # Process the data
        processed_result = {
            "original_data": data,
            "processed_at": datetime.now().isoformat(),
            "word_count": len(data.split()) if isinstance(data, str) else 0,
            "char_count": len(data) if isinstance(data, str) else 0,
            "uppercase": data.upper() if isinstance(data, str) else data
        }
        
        # Store result in database
        conn = get_db_connection()
        if conn:
            try:
                cursor = conn.cursor()
                cursor.execute(
                    """
                    INSERT INTO processing_jobs (job_id, input_data, output_data, status, created_at)
                    VALUES (%s, %s, %s, %s, %s)
                    """,
                    (job_id, json.dumps({"data": data}), json.dumps(processed_result), 'completed', datetime.now())
                )
                conn.commit()
                cursor.close()
            except Exception as e:
                logger.error(f"Database error: {e}")
            finally:
                conn.close()
        
        # Update job status to completed
        redis_client.hset(f"job:{job_id}", mapping={
            "status": "completed",
            "completed_at": datetime.now().isoformat(),
            "result": json.dumps(processed_result)
        })
        
        # Set expiration for job data (24 hours)
        redis_client.expire(f"job:{job_id}", 86400)
        
        logger.info(f"Job {job_id} completed successfully")
        
    except Exception as e:
        logger.error(f"Job {job_id} failed: {e}")
        redis_client.hset(f"job:{job_id}", mapping={
            "status": "failed",
            "error": str(e),
            "failed_at": datetime.now().isoformat()
        })

@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({
        'status': 'healthy',
        'service': 'python-processor',
        'timestamp': datetime.now().isoformat()
    }), 200

@app.route('/api/process', methods=['POST'])
def process_data():
    """Queue a data processing job"""
    try:
        data = request.get_json()
        
        if not data or 'data' not in data:
            return jsonify({'error': 'missing data field'}), 400
        
        # Generate unique job ID
        job_id = str(uuid.uuid4())
        
        # Store initial job status
        redis_client.hset(f"job:{job_id}", mapping={
            "status": "queued",
            "created_at": datetime.now().isoformat(),
            "input_data": json.dumps(data['data'])
        })
        
        # Start background worker
        thread = Thread(target=process_data_worker, args=(job_id, data['data']))
        thread.daemon = True
        thread.start()
        
        return jsonify({
            'message': 'job queued successfully',
            'job_id': job_id,
            'status': 'queued'
        }), 202
        
    except Exception as e:
        logger.error(f"Error queueing job: {e}")
        return jsonify({'error': 'failed to queue job'}), 500

@app.route('/api/jobs/<job_id>', methods=['GET'])
def get_job_status(job_id):
    """Get job status and result"""
    try:
        job_data = redis_client.hgetall(f"job:{job_id}")
        
        if not job_data:
            return jsonify({'error': 'job not found'}), 404
        
        response = {
            'job_id': job_id,
            'status': job_data.get('status'),
            'created_at': job_data.get('created_at')
        }
        
        if 'result' in job_data:
            response['result'] = json.loads(job_data['result'])
        
        if 'error' in job_data:
            response['error'] = job_data['error']
        
        return jsonify(response), 200
        
    except Exception as e:
        logger.error(f"Error fetching job: {e}")
        return jsonify({'error': 'failed to fetch job status'}), 500

@app.route('/api/stats', methods=['GET'])
def get_stats():
    """Get processing statistics"""
    try:
        conn = get_db_connection()
        if not conn:
            return jsonify({'error': 'database connection failed'}), 500
        
        cursor = conn.cursor(cursor_factory=RealDictCursor)
        
        # Get job counts
        cursor.execute("""
            SELECT 
                COUNT(*) as total_jobs,
                COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_jobs,
                COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_jobs
            FROM processing_jobs
        """)
        stats = cursor.fetchone()
        
        cursor.close()
        conn.close()
        
        return jsonify({
            'total_jobs': stats['total_jobs'],
            'completed_jobs': stats['completed_jobs'],
            'failed_jobs': stats['failed_jobs'],
            'timestamp': datetime.now().isoformat()
        }), 200
        
    except Exception as e:
        logger.error(f"Error fetching stats: {e}")
        return jsonify({'error': 'failed to fetch statistics'}), 500

@app.route('/api/housing/import', methods=['POST'])
def import_housing_data():
    """Import housing price index data from CSV"""
    try:
        data = request.get_json()
        
        if not data or 'csv_data' not in data:
            return jsonify({'error': 'missing csv_data field'}), 400
        
        csv_string = data['csv_data']
        
        # Create table if not exists
        conn = get_db_connection()
        if not conn:
            return jsonify({'error': 'database connection failed'}), 500
        
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
            conn.close()
            return jsonify({'error': 'failed to create table'}), 500
        
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
        rows_affected = cursor.rowcount
        conn.commit()
        
        logger.info(f"Successfully imported {len(data_to_insert)} rows, {rows_affected} affected")
        
        cursor.close()
        conn.close()
        
        return jsonify({
            'message': 'data imported successfully',
            'rows_imported': len(data_to_insert),
            'rows_affected': rows_affected
        }), 200
        
    except Exception as e:
        logger.error(f"Error importing housing data: {e}")
        return jsonify({'error': f'failed to import data: {str(e)}'}), 500

@app.route('/api/housing/data', methods=['GET'])
def get_housing_data():
    """Get housing price index data with optional filters"""
    try:
        conn = get_db_connection()
        if not conn:
            return jsonify({'error': 'database connection failed'}), 500
        
        cursor = conn.cursor(cursor_factory=RealDictCursor)
        
        # Get query parameters
        location = request.args.get('location')
        konut_type = request.args.get('type')
        start_date = request.args.get('start_date')
        end_date = request.args.get('end_date')
        
        # Build query
        query = "SELECT * FROM housing_price_index WHERE 1=1"
        params = []
        
        if location:
            query += " AND istanbul_turkiye = %s"
            params.append(location)
        
        if konut_type:
            query += " AND yeni_yeni_olmayan_konut = %s"
            params.append(konut_type)
        
        if start_date:
            query += " AND tarih >= %s"
            params.append(start_date)
        
        if end_date:
            query += " AND tarih <= %s"
            params.append(end_date)
        
        query += " ORDER BY tarih DESC, istanbul_turkiye, yeni_yeni_olmayan_konut"
        
        cursor.execute(query, params)
        results = cursor.fetchall()
        
        cursor.close()
        conn.close()
        
        # Convert Decimal to float for JSON serialization
        for row in results:
            if 'fiyat_endeksi' in row:
                row['fiyat_endeksi'] = float(row['fiyat_endeksi'])
            if 'tarih' in row:
                row['tarih'] = row['tarih'].isoformat()
            if 'created_at' in row:
                row['created_at'] = row['created_at'].isoformat()
            if 'updated_at' in row:
                row['updated_at'] = row['updated_at'].isoformat()
        
        return jsonify({
            'count': len(results),
            'data': results
        }), 200
        
    except Exception as e:
        logger.error(f"Error fetching housing data: {e}")
        return jsonify({'error': 'failed to fetch data'}), 500

if __name__ == '__main__':
    # Test connections on startup
    try:
        redis_client.ping()
        logger.info("Connected to Redis")
    except Exception as e:
        logger.warning(f"Redis connection failed: {e}")
    
    conn = get_db_connection()
    if conn:
        logger.info("Connected to PostgreSQL")
        conn.close()
    else:
        logger.warning("PostgreSQL connection failed")
    
    # Start Flask server
    port = int(os.getenv('PORT', '8081'))
    app.run(host='0.0.0.0', port=port, debug=False)
