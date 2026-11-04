package httpapi

import "testing"

func TestSyncJobStoreAllowsOneRunAtATime(t *testing.T) {
	var store syncJobStore
	first, started := store.begin()
	if !started || first.Status != syncJobRunning {
		t.Fatalf("first job should start running, got started=%v status=%s", started, first.Status)
	}
	again, started := store.begin()
	if started || again.ID != first.ID {
		t.Fatalf("a second request must join the running job, got started=%v id=%s want %s", started, again.ID, first.ID)
	}
	store.finish(first.ID, syncJobDone, "")
	next, started := store.begin()
	if !started || next.ID == first.ID {
		t.Fatal("a new job must be allowed once the previous one finished")
	}
	finished, ok := store.get(first.ID)
	if !ok || finished.Status != syncJobDone || finished.FinishedAt == nil {
		t.Fatalf("finished job should stay readable and be marked done: %+v ok=%v", finished, ok)
	}
}

func TestSyncJobStoreKeepsOnlyRecentJobsAndReturnsCopies(t *testing.T) {
	var store syncJobStore
	var firstID string
	for i := 0; i < syncJobsKept+3; i++ {
		job, _ := store.begin()
		if i == 0 {
			firstID = job.ID
		}
		store.finish(job.ID, syncJobDone, "")
	}
	if _, ok := store.get(firstID); ok {
		t.Error("the oldest finished job should have been dropped")
	}
	job, _ := store.begin()
	store.update(job.ID, func(j *syncJob) { j.Results = append(j.Results, syncResult{TenantName: "A"}) })
	snapshot, _ := store.get(job.ID)
	snapshot.Results[0].TenantName = "changed"
	if again, _ := store.get(job.ID); again.Results[0].TenantName != "A" {
		t.Error("get must return a copy, so a reader cannot change a running job")
	}
}

func TestSyncAllStatusUnknownJobIs404(t *testing.T) {
	var server Server
	if _, ok := server.syncJobs.get("nope"); ok {
		t.Fatal("unknown job must not be found")
	}
}
