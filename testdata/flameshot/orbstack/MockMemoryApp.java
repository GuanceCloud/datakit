import java.util.ArrayList;
import java.util.List;

public class MockMemoryApp {
    public static void main(String[] args) throws Exception {
        int chunkMb = Integer.parseInt(System.getenv().getOrDefault("ALLOC_MB", "8"));
        int steps = Integer.parseInt(System.getenv().getOrDefault("ALLOC_STEPS", "18"));
        int sleepMs = Integer.parseInt(System.getenv().getOrDefault("ALLOC_SLEEP_MS", "500"));

        List<byte[]> blocks = new ArrayList<>();
        System.out.printf("starting allocation: chunkMb=%d steps=%d sleepMs=%d%n", chunkMb, steps, sleepMs);

        for (int i = 0; i < steps; i++) {
            blocks.add(new byte[chunkMb * 1024 * 1024]);
            System.out.printf("allocated chunk %d/%d%n", i + 1, steps);
            Thread.sleep(sleepMs);
        }

        System.out.println("allocation complete, holding memory for observation");
        Thread.sleep(Long.MAX_VALUE);
    }
}
