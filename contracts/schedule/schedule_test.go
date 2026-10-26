package schedule

import (
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

func mustCron(t *testing.T,expr,zone string) Trigger { t.Helper();v,err:=Cron(expr,zone);if err!=nil{t.Fatal(err)};return v }
func TestIntervalSkipsMissedTicks(t *testing.T) {
	tr,err:=Interval(5*time.Second);if err!=nil{t.Fatal(err)}
	anchor:=time.Date(2026,1,1,0,0,0,0,time.UTC);now:=anchor.Add(17*time.Second)
	if got,want:=tr.Next(anchor,now),anchor.Add(20*time.Second);!got.Equal(want){t.Fatalf("Next=%s want %s",got,want)}
	if _,err:=Interval(time.Millisecond*999);err==nil{t.Fatal("accepted short interval")}
}
func TestCronBasicsAndOR(t *testing.T) {
	tr:=mustCron(t,"0 9 * * 1-5","UTC")
	now:=time.Date(2026,10,26,9,0,0,0,time.UTC)
	if got,want:=tr.Next(time.Time{},now),time.Date(2026,10,27,9,0,0,0,time.UTC);!got.Equal(want){t.Fatalf("got %s want %s",got,want)}
	tr=mustCron(t,"0 0 1 * 1","UTC")
	now=time.Date(2026,10,2,0,0,0,0,time.UTC)
	if got,want:=tr.Next(time.Time{},now),time.Date(2026,10,5,0,0,0,0,time.UTC);!got.Equal(want){t.Fatalf("OR got %s want %s",got,want)}
}
func TestDST(t *testing.T) {
	tr:=mustCron(t,"30 2 * * *","America/New_York")
	now:=time.Date(2026,3,8,0,0,0,0,time.UTC)
	if got,want:=tr.Next(time.Time{},now),time.Date(2026,3,9,6,30,0,0,time.UTC);!got.Equal(want){t.Fatalf("spring got %s want %s",got,want)}
	tr=mustCron(t,"30 1 * * *","America/New_York")
	first:=time.Date(2026,11,1,4,0,0,0,time.UTC)
	if got,want:=tr.Next(time.Time{},first),time.Date(2026,11,1,5,30,0,0,time.UTC);!got.Equal(want){t.Fatalf("first got %s want %s",got,want)}
	if got,want:=tr.Next(time.Time{},time.Date(2026,11,1,5,30,0,0,time.UTC)),time.Date(2026,11,1,6,30,0,0,time.UTC);!got.Equal(want){t.Fatalf("second got %s want %s",got,want)}
}
func TestCronRejectsInvalid(t *testing.T) {
	for _,expr:=range []string{"","@daily","0 0 0 1 1 *","JAN 1 * * *","0 0 2-1 * *","*/0 * * * *","0 0 30 2 *"} {
		t.Run(expr,func(t *testing.T){if _,err:=Cron(expr,"UTC");err==nil{t.Fatalf("accepted %q",expr)}})
	}
	if _,err:=Cron("* * * * *","Local");err==nil{t.Fatal("accepted Local")}
	if _,err:=Cron("* * * * *","Invalid/Zone");err==nil{t.Fatal("accepted invalid zone")}
}
func TestFromSpec(t *testing.T) {
	if _,err:=FromSpec(&pb.ScheduleSpec{});err==nil{t.Fatal("accepted empty trigger")}
	tr,err:=FromSpec(&pb.ScheduleSpec{Trigger:&pb.ScheduleSpec_IntervalMs{IntervalMs:1000}});if err!=nil||tr.kind!=intervalKind{t.Fatalf("trigger=%#v err=%v",tr,err)}
	tr,err=FromSpec(&pb.ScheduleSpec{Trigger:&pb.ScheduleSpec_Cron{Cron:"0 0 * * *"},Timezone:"UTC"});if err!=nil||tr.kind!=cronKind{t.Fatalf("trigger=%#v err=%v",tr,err)}
}
