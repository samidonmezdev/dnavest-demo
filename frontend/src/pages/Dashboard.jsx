import { useEffect, useState } from 'react';
import apiClient from '../api/client';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, BarChart, Bar } from 'recharts';
import { useAuth } from '../context/AuthContext';
import { LogOut, User } from 'lucide-react';

export default function Dashboard() {
    const { user, logout } = useAuth();
    const [stats, setStats] = useState(null);
    const [chartData, setChartData] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [statsRes, chartsRes] = await Promise.all([
                    apiClient.get('/data/stats'),
                    apiClient.get('/data/housing/charts')
                ]);
                setStats(statsRes.data);
                setChartData(chartsRes.data);
            } catch (error) {
                console.error('Error fetching dashboard data:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, []);

    if (loading) return <div className="loading">Loading Dashboard...</div>;

    return (
        <div className="dashboard-container">
            <header className="dashboard-header">
                <h1>Finscope Dashboard</h1>
                <div className="user-info">
                    <span><User size={16} /> Welcome, {user?.name || user?.email}</span>
                    <button onClick={logout} className="logout-btn">
                        <LogOut size={16} /> Logout
                    </button>
                </div>
            </header>

            <main className="dashboard-content">
                <div className="stats-grid">
                    <div className="stat-card">
                        <h3>Total Users</h3>
                        <p className="stat-value">{stats?.total_users || 0}</p>
                    </div>
                    <div className="stat-card">
                        <h3>System Status</h3>
                        <p className="stat-value">Active</p>
                    </div>
                </div>

                <div className="charts-section">
                    <h2>Housing Market Trends</h2>

                    <div className="chart-container">
                        <h3>Price Trends (TRY)</h3>
                        <ResponsiveContainer width="100%" height={300}>
                            {/* Assuming chartData structure matches expected housing data */}
                            <LineChart data={chartData}>
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="date" />
                                <YAxis />
                                <Tooltip />
                                <Legend />
                                <Line type="monotone" dataKey="price_avg" stroke="#8884d8" name="Avg Price" />
                            </LineChart>
                        </ResponsiveContainer>
                    </div>

                    <div className="chart-container">
                        <h3>Sales Volume</h3>
                        <ResponsiveContainer width="100%" height={300}>
                            <BarChart data={chartData}>
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="date" />
                                <YAxis />
                                <Tooltip />
                                <Legend />
                                <Bar dataKey="sales_count" fill="#82ca9d" name="Sales Count" />
                            </BarChart>
                        </ResponsiveContainer>
                    </div>
                </div>
            </main>
        </div>
    );
}
