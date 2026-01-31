import os
import logging
from datetime import datetime
from flask import Flask, jsonify

import psycopg2


# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

app = Flask(__name__)


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

@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({
        'status': 'healthy',
        'service': 'python-processor',
        'timestamp': datetime.now().isoformat()
    }), 200

if __name__ == '__main__':
    # Test DB connection
    conn = get_db_connection()

    if conn:
        logger.info("Connected to PostgreSQL")
        conn.close()
    else:
        logger.warning("PostgreSQL connection failed")

    # Auto-import housing data on startup
    try:
        from import_housing_data import import_csv_data
        csv_file = 'housing_data.csv'
        if os.path.exists(csv_file):
            logger.info(f"Attempting to import {csv_file} on startup...")
            if import_csv_data(csv_file):
                logger.info("Startup data import completed successfully")
            else:
                logger.error("Startup data import failed")
        else:
            logger.warning(f"{csv_file} not found, skipping startup import")
    except Exception as e:
        logger.error(f"Error during startup data import: {e}")
    
    # Start Flask server
    port = int(os.getenv('PORT', '8081'))
    app.run(host='0.0.0.0', port=port, debug=False)
