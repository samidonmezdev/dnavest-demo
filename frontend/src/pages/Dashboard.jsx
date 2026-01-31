import { useEffect, useState } from 'react';
import apiClient from '../api/client';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, BarChart, Bar } from 'recharts';
import { useAuth } from '../context/AuthContext';
import { LogOut, User } from 'lucide-react';

export default function Dashboard() {
    const { user, logout } = useAuth();
    const [stats, setStats] = useState(null);
    const [chartData, setChartData] = useState([]);
    const [processedData, setProcessedData] = useState([]);
    const [comparisonData, setComparisonData] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [statsRes, chartsRes] = await Promise.all([
                    apiClient.get('/data/stats'),
                    apiClient.get('/data/housing/charts')
                ]);
                setStats(statsRes.data);
                const rawData = chartsRes.data.data || [];
                setChartData(rawData);

                // Process data for multi-line chart
                const groupedByDate = rawData.reduce((acc, curr) => {
                    const date = new Date(curr.tarih).toLocaleDateString();
                    if (!acc[date]) {
                        acc[date] = {
                            date,
                            originalDate: curr.tarih
                        };
                    }
                    // Create keys like "Türkiye_Yeni Konut", "İstanbul_Yeni Olmayan Konut"
                    const key = `${curr.istanbul_turkiye}_${curr.yeni_yeni_olmayan_konut}`;
                    acc[date][key] = curr.fiyat_endeksi;
                    return acc;
                }, {});

                const processed = Object.values(groupedByDate).sort((a, b) => new Date(a.originalDate) - new Date(b.originalDate));
                setProcessedData(processed);

                // Prepare comparison data (latest available date)
                if (processed.length > 0) {
                    const latest = processed[processed.length - 1];
                    const compData = [
                        { name: 'TR-New', value: latest['Türkiye_Yeni Konut'], fill: '#8884d8' },
                        { name: 'TR-Old', value: latest['Türkiye_Yeni Olmayan Konut'], fill: '#82ca9d' },
                        { name: 'IST-New', value: latest['İstanbul_Yeni Konut'], fill: '#ffc658' },
                        { name: 'IST-Old', value: latest['İstanbul_Yeni Olmayan Konut'], fill: '#ff7300' },
                    ].filter(item => item.value !== undefined);
                    setComparisonData(compData);
                }

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
                    <div className="stat-card">
                        <h3>Total Data Points</h3>
                        <p className="stat-value">{chartData.length}</p>
                    </div>
                </div>

                <div className="charts-section" style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
                    <h2 style={{ width: '100%', marginBottom: '15px' }}>Housing Market Trends</h2>

                    {processedData.length > 0 ? (
                        <>
                            <div className="charts-grid" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px' }}>
                                <div className="chart-container">
                                    <h3>Turkey Trends</h3>
                                    <ResponsiveContainer width="100%" height={300}>
                                        <LineChart data={processedData}>
                                            <CartesianGrid strokeDasharray="3 3" />
                                            <XAxis dataKey="date" />
                                            <YAxis domain={['auto', 'auto']} />
                                            <Tooltip />
                                            <Legend />
                                            <Line type="monotone" dataKey="Türkiye_Yeni Konut" name="TR New" stroke="#8884d8" strokeWidth={2} dot={false} />
                                            <Line type="monotone" dataKey="Türkiye_Yeni Olmayan Konut" name="TR Old" stroke="#82ca9d" strokeWidth={2} dot={false} />
                                        </LineChart>
                                    </ResponsiveContainer>
                                </div>

                                <div className="chart-container">
                                    <h3>Istanbul Trends</h3>
                                    <ResponsiveContainer width="100%" height={300}>
                                        <LineChart data={processedData}>
                                            <CartesianGrid strokeDasharray="3 3" />
                                            <XAxis dataKey="date" />
                                            <YAxis domain={['auto', 'auto']} />
                                            <Tooltip />
                                            <Legend />
                                            <Line type="monotone" dataKey="İstanbul_Yeni Konut" name="IST New" stroke="#ffc658" strokeWidth={2} dot={false} />
                                            <Line type="monotone" dataKey="İstanbul_Yeni Olmayan Konut" name="IST Old" stroke="#ff7300" strokeWidth={2} dot={false} />
                                        </LineChart>
                                    </ResponsiveContainer>
                                </div>
                            </div>

                            <div className="charts-row">
                                <div className="chart-container">
                                    <h3>Latest Snapshot</h3>
                                    <ResponsiveContainer width="100%" height={300}>
                                        <BarChart data={comparisonData}>
                                            <CartesianGrid strokeDasharray="3 3" />
                                            <XAxis dataKey="name" />
                                            <YAxis />
                                            <Tooltip />
                                            <Bar dataKey="value" name="Price Index" />
                                        </BarChart>
                                    </ResponsiveContainer>
                                </div>
                            </div>

                            <div className="data-table-container" style={{ background: 'white', padding: '20px', borderRadius: '8px', boxShadow: '0 2px 4px rgba(0,0,0,0.1)' }}>
                                <h3>Detailed Market Data</h3>
                                <div className="table-wrapper" style={{ maxHeight: '400px', overflowY: 'auto' }}>
                                    <table style={{ width: '100%', borderCollapse: 'collapse', marginTop: '10px' }}>
                                        <thead>
                                            <tr style={{ background: '#f8f9fa', textAlign: 'left' }}>
                                                <th style={{ padding: '12px', borderBottom: '2px solid #dee2e6' }}>Date</th>
                                                <th style={{ padding: '12px', borderBottom: '2px solid #dee2e6' }}>Location</th>
                                                <th style={{ padding: '12px', borderBottom: '2px solid #dee2e6' }}>Type</th>
                                                <th style={{ padding: '12px', borderBottom: '2px solid #dee2e6' }}>Price Index</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {chartData.map((row, index) => (
                                                <tr key={index} style={{ borderBottom: '1px solid #eee' }}>
                                                    <td style={{ padding: '10px' }}>{new Date(row.tarih).toLocaleDateString()}</td>
                                                    <td style={{ padding: '10px' }}>{row.istanbul_turkiye}</td>
                                                    <td style={{ padding: '10px' }}>{row.yeni_yeni_olmayan_konut}</td>
                                                    <td style={{ padding: '10px', fontWeight: 'bold' }}>{row.fiyat_endeksi}</td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </>
                    ) : (
                        <div className="no-data">No data available to display charts.</div>
                    )}
                </div>
            </main >
        </div >
    );
}
