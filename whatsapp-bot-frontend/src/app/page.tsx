import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { MessageSquare, Users, Zap, Shield, ArrowRight } from 'lucide-react';

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      {/* Header */}
      <header className="border-b bg-white/80 backdrop-blur-sm">
        <div className="container mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <MessageSquare className="h-8 w-8 text-blue-600" />
            <h1 className="text-2xl font-bold text-gray-900">WhatsApp Bot Notify</h1>
          </div>
          <div className="flex items-center space-x-4">
            <Link href="/docs">
              <Button variant="ghost">Dokumentasi</Button>
            </Link>
            <Link href="/login">
              <Button>Login</Button>
            </Link>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="py-20">
        <div className="container mx-auto px-4 text-center">
          <h2 className="text-5xl font-bold text-gray-900 mb-6">
            Sistem Bot WhatsApp
            <span className="text-blue-600"> Multi-User</span>
          </h2>
          <p className="text-xl text-gray-600 mb-8 max-w-3xl mx-auto">
            Platform canggih untuk mengelola bot WhatsApp dengan dukungan multi-user, 
            API yang powerful, dan interface yang intuitif untuk automasi pesan.
          </p>
          <div className="flex items-center justify-center space-x-4">
            <Link href="/login">
              <Button size="lg" className="text-lg px-8 py-3">
                Mulai Sekarang
                <ArrowRight className="ml-2 h-5 w-5" />
              </Button>
            </Link>
            <Link href="/docs">
              <Button variant="outline" size="lg" className="text-lg px-8 py-3">
                Lihat Dokumentasi
              </Button>
            </Link>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 bg-white">
        <div className="container mx-auto px-4">
          <h3 className="text-3xl font-bold text-center text-gray-900 mb-12">
            Fitur Unggulan
          </h3>
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            <Card>
              <CardHeader>
                <Users className="h-12 w-12 text-blue-600 mb-4" />
                <CardTitle>Multi-User Support</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Setiap pengguna memiliki instance bot yang terpisah dan terisolasi
                  untuk keamanan dan privasi maksimal.
                </CardDescription>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <Zap className="h-12 w-12 text-blue-600 mb-4" />
                <CardTitle>Real-time Updates</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Monitoring status koneksi dan aktivitas bot secara real-time
                  dengan WebSocket dan notifikasi instant.
                </CardDescription>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <MessageSquare className="h-12 w-12 text-blue-600 mb-4" />
                <CardTitle>API Lengkap</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  RESTful API yang komprehensif untuk mengirim pesan personal,
                  grup, dan broadcast dengan dokumentasi lengkap.
                </CardDescription>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <Shield className="h-12 w-12 text-blue-600 mb-4" />
                <CardTitle>Keamanan Tinggi</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Autentikasi JWT, rate limiting, dan isolasi data per pengguna
                  untuk menjamin keamanan sistem.
                </CardDescription>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 bg-blue-600">
        <div className="container mx-auto px-4 text-center">
          <h3 className="text-3xl font-bold text-white mb-6">
            Siap Memulai Automasi WhatsApp?
          </h3>
          <p className="text-xl text-blue-100 mb-8 max-w-2xl mx-auto">
            Bergabunglah dengan platform bot WhatsApp terdepan dan mulai 
            otomatisasi pesan Anda hari ini.
          </p>
          <Link href="/login">
            <Button size="lg" variant="secondary" className="text-lg px-8 py-3">
              Akses Dashboard
              <ArrowRight className="ml-2 h-5 w-5" />
            </Button>
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-white py-12">
        <div className="container mx-auto px-4">
          <div className="grid md:grid-cols-3 gap-8">
            <div>
              <div className="flex items-center space-x-2 mb-4">
                <MessageSquare className="h-6 w-6" />
                <span className="text-lg font-semibold">WhatsApp Bot Notify</span>
              </div>
              <p className="text-gray-400">
                Platform automasi WhatsApp yang powerful dan mudah digunakan
                untuk kebutuhan bisnis modern.
              </p>
            </div>
            <div>
              <h4 className="text-lg font-semibold mb-4">Fitur</h4>
              <ul className="space-y-2 text-gray-400">
                <li>Multi-User Support</li>
                <li>Real-time Monitoring</li>
                <li>RESTful API</li>
                <li>Broadcast Messages</li>
              </ul>
            </div>
            <div>
              <h4 className="text-lg font-semibold mb-4">Bantuan</h4>
              <ul className="space-y-2 text-gray-400">
                <li>
                  <Link href="/docs" className="hover:text-white transition-colors">
                    Dokumentasi API
                  </Link>
                </li>
                <li>
                  <Link href="/login" className="hover:text-white transition-colors">
                    Login Dashboard
                  </Link>
                </li>
              </ul>
            </div>
          </div>
          <div className="border-t border-gray-800 mt-8 pt-8 text-center text-gray-400">
            <p>&copy; 2024 WhatsApp Bot Notify. All rights reserved.</p>
          </div>
        </div>
      </footer>
    </div>
  );
}