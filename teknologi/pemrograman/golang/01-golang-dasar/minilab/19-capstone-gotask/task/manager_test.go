package task

import (
	"testing"
	"time"
)

func newDeadline(hoursFromNow float64) time.Time {
	return time.Now().Add(time.Duration(hoursFromNow * float64(time.Hour)))
}

func TestManager_AddTask(t *testing.T) {
	m := NewManager()
	deadline := newDeadline(48)

	task, err := m.AddTask("Belajar Struct", deadline, "go")
	if err != nil {
		t.Fatalf("AddTask error: %v", err)
	}
	if task.ID != 1 {
		t.Errorf("ID pertama = %d; want 1", task.ID)
	}

	task2, _ := m.AddTask("Belajar Interface", deadline, "go")
	if task2.ID != 2 {
		t.Errorf("ID kedua = %d; want 2", task2.ID)
	}
}

func TestManager_AddTask_TitleKosong(t *testing.T) {
	m := NewManager()
	_, err := m.AddTask("", time.Now().Add(time.Hour), "")
	if err == nil {
		t.Error("harus error jika title kosong")
	}
}

func TestManager_CompleteTask(t *testing.T) {
	m := NewManager()
	task, _ := m.AddTask("Test", newDeadline(24), "")

	err := m.CompleteTask(task.ID)
	if err != nil {
		t.Fatalf("CompleteTask error: %v", err)
	}

	tasks := m.ListTasks("")
	if tasks[0].Status != StatusDone {
		t.Errorf("Status = %q; want done", tasks[0].Status)
	}
}

func TestManager_CompleteTask_TidakAda(t *testing.T) {
	m := NewManager()
	err := m.CompleteTask(999)
	if err == nil {
		t.Error("harus error jika ID tidak ada")
	}
}

func TestManager_ListTasks(t *testing.T) {
	m := NewManager()
	m.AddTask("Task 1", newDeadline(24), "go")
	m.AddTask("Task 2", newDeadline(48), "belajar")
	m.AddTask("Task 3", newDeadline(72), "go")

	all := m.ListTasks("")
	if len(all) != 3 {
		t.Errorf("len(ListTasks(\"\")) = %d; want 3", len(all))
	}

	goTasks := m.ListTasks("go")
	if len(goTasks) != 2 {
		t.Errorf("len(ListTasks(\"go\")) = %d; want 2", len(goTasks))
	}
}

func TestManager_DeleteTask(t *testing.T) {
	m := NewManager()
	task, _ := m.AddTask("Hapus ini", newDeadline(24), "")

	err := m.DeleteTask(task.ID)
	if err != nil {
		t.Fatalf("DeleteTask error: %v", err)
	}

	all := m.ListTasks("")
	if len(all) != 0 {
		t.Errorf("task seharusnya sudah dihapus, len = %d", len(all))
	}
}

func TestManager_OverdueTasks(t *testing.T) {
	m := NewManager()
	m.AddTask("Sudah lewat", newDeadline(-24), "")
	m.AddTask("Belum lewat", newDeadline(24), "")
	m.AddTask("Sudah selesai", newDeadline(-24), "")

	tasks := m.ListTasks("")
	m.CompleteTask(tasks[2].ID)

	overdue := m.OverdueTasks()
	if len(overdue) != 1 {
		t.Errorf("OverdueTasks() = %d; want 1", len(overdue))
	}
}
