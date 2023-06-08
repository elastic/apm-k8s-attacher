package co.elastic.apm.example.test;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.InetSocketAddress;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;

public class SimpleMockApmServer {
    public static final int PORT = 8027; //0 would result in a random assigned port
    private static volatile HttpServer TheServerInstance;

    public static void main(String[] args) throws Exception {
        SimpleMockApmServer server = new SimpleMockApmServer();
        System.out.println("PORT: "+server.start());
    }

    public synchronized int start() throws IOException {
        if (TheServerInstance != null) {
            throw new IOException("Ooops, you can't start this instance more than once");
        }
        InetSocketAddress addr = new InetSocketAddress("0.0.0.0", PORT);
        HttpServer server = HttpServer.create(addr, 10);
        server.createContext("/", new RootHandler());

        server.start();
        TheServerInstance = server;
        System.out.println("MockApmServer started on port " + server.getAddress().getPort());
        return server.getAddress().getPort();
    }

    class RootHandler implements HttpHandler {
        public void handle(HttpExchange t) {
            try {
                InputStream body = t.getRequestBody();
                ByteArrayOutputStream bytes = new ByteArrayOutputStream();
                byte[] buffer = new byte[8 * 1024];
                int lengthRead;
                while ((lengthRead = body.read(buffer)) > 0) {
                    bytes.write(buffer, 0, lengthRead);
                }
                String[] lines = bytes.toString().split("[\r\n]");
                for (String line : lines) {
                    System.out.println(line);
                }
                String response = "{}";
                if (t.getRequestURI().toString().equals("/")) {
                    response = "{\"build_date\": \"2023-02-13T13:01:54Z\", \"build_sha\": \"8638b035d700e5e85e376252402b5375e4d4190b\", \"publish_ready\": true, \"version\": \"8.6.2\"}";
                }
                t.sendResponseHeaders(200, response.length());
                OutputStream os = t.getResponseBody();
                os.write(response.getBytes());
                os.close();
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
    }

}
